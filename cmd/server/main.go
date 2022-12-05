package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/compute/metadata"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/Silicon-Ally/gqlerr"
	"github.com/Silicon-Ally/silicon-starter/authn/fireauth"
	"github.com/Silicon-Ally/silicon-starter/authn/session"
	"github.com/Silicon-Ally/silicon-starter/cmd/server/generated"
	"github.com/Silicon-Ally/silicon-starter/cmd/server/graph"
	"github.com/Silicon-Ally/silicon-starter/common/flagext"
	"github.com/Silicon-Ally/silicon-starter/db/sqldb"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/namsral/flag"
	"github.com/rs/cors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/api/option"

	firebase "firebase.google.com/go/v4"
)

func main() {
	if err := run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return errors.New("args cannot be empty")
	}

	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
	var (
		minLogLevel zapcore.Level = zapcore.WarnLevel

		localDSN = fs.String("local_dsn", "", "If set, override the DB addresses retrieved from the sops configuration. Can only be used when running locally.")

		sopsConfigPath = fs.String("sops_encrypted_config", "", "A JSON-formatted configuration file for our main server, parseable by the SOPS tool (https://github.com/mozilla/sops).")
		port           = fs.Int("port", 8080, "The port to serve the backend's HTTP service on.")
		projectID      = fs.String("project_id", "", "The GCP project ID this service runs in/as. Only set in deployed environments.")

		debug = fs.Bool("debug", false, "If true, enable the /playground endpoint for testing out GraphQL queries and CORS debugging.")

		allowedCORSOrigins flagext.StringList
	)
	fs.Var(&minLogLevel, "min_log_level", "If set, retains logs at the given level and above. Options: 'debug', 'info', 'warn', 'error', 'dpanic', 'panic', 'fatal' - default warn.")
	fs.Var(&allowedCORSOrigins, "allowed_cors_origins", "A comma-delimited list of origins to allow for CORS (Cross-Origin Resource Sharing).")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Allows for passing in configuration via a -config path/to/env-file.conf
	// flag, see https://pkg.go.dev/github.com/namsral/flag#readme-usage
	fs.String(flag.DefaultConfigFlagname, "", "path to config file")
	if err := fs.Parse(args[1:]); err != nil {
		return fmt.Errorf("failed to parse flags: %v", err)
	}

	if *localDSN != "" && metadata.OnGCE() {
		return errors.New("--local_dsn set outside of local environment")
	}

	var config zap.Config
	if *debug {
		config = zap.NewDevelopmentConfig()
	} else {
		config = zap.NewProductionConfig()
	}
	config.Level = zap.NewAtomicLevelAt(minLogLevel)
	logger, err := config.Build()
	if err != nil {
		return fmt.Errorf("failed to init logger: %w", err)
	}

	requiredFlags := []struct {
		flagName string
		val      *string
	}{
		{"sops_encrypted_config", sopsConfigPath},
		{"project_id", projectID},
	}
	for _, rf := range requiredFlags {
		if *rf.val == "" {
			return fmt.Errorf("no --%s was specified", rf.flagName)
		}
	}

	// We use sops for secret management. That isn't necessary (or possible) to include
	// out of the box working in this repo.
	// See the documentation in secrets/README.md for more information.
	// logger.Info("Decrypting configuration", zap.String("sops_path", *sopsConfigPath))
	// cfg, err := secrets.Load(*sopsConfigPath)
	// if err != nil {
	//	return fmt.Errorf("failed to decrypt configuration: %w", err)
	// }

	var postgresCfg *pgxpool.Config
	if *localDSN != "" {
		if postgresCfg, err = pgxpool.ParseConfig(*localDSN); err != nil {
			return fmt.Errorf("failed to parse local DSN: %w", err)
		}
	}

	logger.Info("Connecting to database", zap.String("db_host", postgresCfg.ConnConfig.Host))
	pgConn, err := pgxpool.ConnectConfig(ctx, postgresCfg)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	logger.Info("Pinging database")
	if err := pgConn.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}
	db, err := sqldb.New(pgConn)
	if err != nil {
		return fmt.Errorf("failed to init sqldb: %w", err)
	}

	logger.Info("Initializing Firebase Connection")
	// Without option.WithQuotaProject(...), the service will authenticate using
	// your default project, which may not have the
	// identitytoolkit.googleapis.com service enabled.
	firebaseApp, err := firebase.NewApp(
		ctx,
		&firebase.Config{ProjectID: *projectID},
		option.WithQuotaProject(*projectID))
	if err != nil {
		return fmt.Errorf("failed to init Firebase client: %w", err)
	}
	firebaseAuth, err := firebaseApp.Auth(ctx)
	if err != nil {
		return fmt.Errorf("failed to init Firebase auth client: %w", err)
	}

	logger.Info("Initializing GraphQL resolvers")
	resolver, err := graph.NewResolver(&graph.ResolverConfig{
		DB:     db,
		Logger: logger,
	})
	if err != nil {
		return fmt.Errorf("failed to init resolver: %w", err)
	}

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: resolver}))
	srv.SetErrorPresenter(gqlerr.ErrorPresenter(logger))

	mux := http.NewServeMux()

	if *debug {
		// Enable the playground endpoint
		mux.Handle("/api/playground", playground.Handler("GraphQL playground", "/api/graphql"))
		logger.Info(
			"Running in debug mode, adding playground handler",
			zap.String("playground_url", fmt.Sprintf("http://localhost:%d/api/playground", *port)),
		)
	}

	sess := session.New(
		fireauth.New(firebaseAuth),
		db,
		logger.With(zap.Namespace("firebase auth")),
	)

	mux.Handle("/api/graphql", srv)
	mux.Handle("/api/sessionLogin", sess.LoginHandler())
	mux.Handle("/api/sessionLogout", sess.LogoutHandler())

	handler := sess.WithAuthorization(mux, "/api/sessionLogin")
	handler = withCORS(handler, []string(allowedCORSOrigins), *debug, logger.With(zap.Namespace("cors")))

	addr := fmt.Sprintf(":%d", *port)
	logger.Info("Starting server", zap.String("server_addr", addr))

	// Created with https://textkool.com/en/ascii-art-generator?hl=default&vl=default&font=Pagga&text=SILICON%0ASTARTER
	fmt.Println()
	fmt.Println(`
░█▀▀░▀█▀░█░░░▀█▀░█▀▀░█▀█░█▀█
░▀▀█░░█░░█░░░░█░░█░░░█░█░█░█
░▀▀▀░▀▀▀░▀▀▀░▀▀▀░▀▀▀░▀▀▀░▀░▀
░█▀▀░▀█▀░█▀█░█▀▄░▀█▀░█▀▀░█▀▄
░▀▀█░░█░░█▀█░█▀▄░░█░░█▀▀░█▀▄
░▀▀▀░░▀░░▀░▀░▀░▀░░▀░░▀▀▀░▀░▀`)
	fmt.Println()

	if err := http.ListenAndServe(addr, handler); err != nil {
		return fmt.Errorf("http.ListenAndServe: %w", err)
	}

	return nil
}

func withCORS(next http.Handler, allowedOrigins []string, debug bool, logger *zap.Logger) http.Handler {
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowCredentials: true,
		Debug:            debug,
		AllowedMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		},
		// We might want to be more selective in the future, but receiving extra
		// headers isn't a big deal.
		AllowedHeaders: []string{"*"},
	})
	corsHandler.Log = &corsLogger{logger.Sugar()}

	return corsHandler.Handler(next)
}

// corsLogger is a thin shim around our zap-based logging to match the
// cors.Logger interface.
type corsLogger struct {
	logger *zap.SugaredLogger
}

func (c *corsLogger) Printf(format string, args ...interface{}) {
	c.logger.Infof(format, args...)
}

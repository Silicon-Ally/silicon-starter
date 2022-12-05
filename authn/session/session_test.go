package session

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Silicon-Ally/silicon-starter/authn"
	"github.com/Silicon-Ally/silicon-starter/testing/testdb"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"go.uber.org/zap/zaptest"
)

func TestLoginHandler(t *testing.T) {
	// These functions stub out time.Since, always returning the delta from a
	// fixed point in time.
	now := time.Unix(123456789, 0)
	since := func(t time.Time) time.Duration {
		return now.Sub(t)
	}

	validToken := &authn.Token{
		UserInfo: &authn.UserInfo{
			UserID:       "user-id",
			Email:        "test@example.com",
			AuthProvider: authn.Google,
		},
		AuthTime: now.Add(-5 * time.Second),
	}
	fromValid := func(modify func(tkn *authn.Token)) string {
		// Start with a valid token.
		tkn := &authn.Token{
			UserInfo: &authn.UserInfo{
				UserID:       "user-id",
				Email:        "test@example.com",
				AuthProvider: authn.Google,
			},
			AuthTime: now.Add(-5 * time.Second),
		}
		modify(tkn)
		return encodeAuthToken(t, tkn)
	}

	validSessionCookie := &sessionCookie{
		ExpiresIn: time.Hour * 24 * 14,
		Token:     validToken,
	}

	tests := []struct {
		desc        string
		req         *LoginRequest
		wantStatus  int
		wantCookies []*http.Cookie
	}{
		{
			desc: "valid session",
			req: &LoginRequest{
				IDToken:   fromValid(func(*authn.Token) {} /* noop */),
				CSRFToken: "not-used-currently",
			},
			wantStatus: http.StatusOK,
			wantCookies: []*http.Cookie{
				{
					Name:     "__session",
					Path:     "/",
					Value:    encodeSessionCookie(t, validSessionCookie),
					MaxAge:   60 * 60 * 24 * 14, // 14 days, in seconds
					HttpOnly: true,
					Secure:   true,
					SameSite: http.SameSiteStrictMode,
				},
			},
		},
		{
			desc: "valid session, set name",
			req: &LoginRequest{
				Name:      "A New Test User",
				IDToken:   fromValid(func(*authn.Token) {} /* noop */),
				CSRFToken: "not-used-currently",
			},
			wantStatus: http.StatusOK,
			wantCookies: []*http.Cookie{
				{
					Name:     "__session",
					Path:     "/",
					Value:    encodeSessionCookie(t, validSessionCookie),
					MaxAge:   60 * 60 * 24 * 14, // 14 days, in seconds
					HttpOnly: true,
					Secure:   true,
					SameSite: http.SameSiteStrictMode,
				},
			},
		},
		{
			desc: "invalid session, token too old",
			req: &LoginRequest{
				IDToken:   fromValid(func(tkn *authn.Token) { tkn.AuthTime = now.Add(-6 * time.Minute) }),
				CSRFToken: "not-used-currently",
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			desc: "invalid session, token fails verification",
			req: &LoginRequest{
				IDToken:   "this won't pass verification",
				CSRFToken: "not-used-currently",
			},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			fAuth := &fakeAuth{}
			tdb := testdb.New()
			sess := New(
				fAuth,
				tdb,
				zaptest.NewLogger(t),
			)
			sess.since = since

			ts := httptest.NewServer(sess.LoginHandler())
			defer ts.Close()

			body := strings.NewReader(encodeLoginRequest(t, test.req))
			req, err := http.NewRequest(http.MethodPost, ts.URL, body)
			if err != nil {
				t.Fatalf("http.NewRequest: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")
			resp, err := ts.Client().Do(req)
			if err != nil {
				t.Fatalf("failed to issue login request: %v", err)
			}

			if resp.StatusCode != test.wantStatus {
				t.Errorf("login response code was %d, want %d", resp.StatusCode, test.wantStatus)
			}
			if diff := cmp.Diff(test.wantCookies, resp.Cookies(), cookieDiffOpts()); diff != "" {
				t.Errorf("unexpected cookies in login response (-want +got)\n%s", diff)
			}
		})
	}
}

func cookieDiffOpts() cmp.Option {
	return cmp.Options{
		// Ignore the 'Raw' parameter of cookies, because it's just noise and
		// redundant with the structured contents.
		cmpopts.IgnoreFields(http.Cookie{}, "Raw"),
		cmpopts.EquateEmpty(),
	}
}

// fakeAuth implements the session.Auth interface for use in tests. Any ID
// token that is a JSON-formatted authn.Token struct will pass 'verification'. A
// session cookie is the same thing, but wrapped in the sessionCookie struct.
type fakeAuth struct {
	revoked []authn.UserID
}

type sessionCookie struct {
	ExpiresIn time.Duration
	Token     *authn.Token
}

func (f *fakeAuth) VerifyIDToken(ctx context.Context, idToken string) (*authn.Token, error) {
	return parseToken(idToken)
}

func parseToken(idToken string) (*authn.Token, error) {
	var tkn authn.Token
	if err := json.NewDecoder(strings.NewReader(idToken)).Decode(&tkn); err != nil {
		return nil, fmt.Errorf("failed to decode id token: %w", err)
	}
	return &tkn, nil
}

func (f *fakeAuth) SessionCookie(ctx context.Context, idToken string, expiresIn time.Duration) (string, error) {
	tkn, err := parseToken(idToken)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&sessionCookie{
		ExpiresIn: expiresIn,
		Token:     tkn,
	}); err != nil {
		return "", fmt.Errorf("failed to encode session cookie: %w", err)
	}

	// session.Auth expects that SessionCookie returns a string that is a valid
	// cookie value, which a JSON-encoded value wouldn't be, so we escape it.
	return url.QueryEscape(buf.String()), nil
}

func (f *fakeAuth) VerifySessionCookie(ctx context.Context, sessionCookie string) (*authn.Token, error) {
	sc, err := parseSessionCookie(sessionCookie)
	if err != nil {
		return nil, err
	}
	return sc.Token, nil
}

func parseSessionCookie(cookieStr string) (*sessionCookie, error) {
	var sc sessionCookie
	if err := json.NewDecoder(strings.NewReader(cookieStr)).Decode(&sc); err != nil {
		return nil, fmt.Errorf("failed to decode session cookie: %w", err)
	}
	return &sc, nil
}

func (f *fakeAuth) RevokeRefreshTokens(ctx context.Context, uID authn.UserID) error {
	f.revoked = append(f.revoked, uID)
	return nil
}

func encodeAuthToken(t *testing.T, tkn *authn.Token) string {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(tkn); err != nil {
		t.Fatalf("failed to encode auth token: %v", err)
	}
	return buf.String()
}

func encodeSessionCookie(t *testing.T, sc *sessionCookie) string {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(sc); err != nil {
		t.Fatalf("failed to encode session cookie: %v", err)
	}
	return url.QueryEscape(buf.String())
}

func encodeLoginRequest(t *testing.T, req *LoginRequest) string {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(req); err != nil {
		t.Fatalf("failed to encode login request: %v", err)
	}
	return buf.String()
}

func encodeUserInfo(t *testing.T, ui *authn.UserInfo) string {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(ui); err != nil {
		t.Fatalf("failed to encode user info: %v", err)
	}
	return url.QueryEscape(buf.String())
}

func TestParseLoginRequest(t *testing.T) {
	tests := []struct {
		desc    string
		in      string
		want    *LoginRequest
		wantErr bool
	}{
		{
			desc: "valid request",
			in:   `{"idToken": "test token", "csrfToken": "csrf"}`,
			want: &LoginRequest{
				IDToken:   "test token",
				CSRFToken: "csrf",
			},
		},
		{
			desc: "valid request with name",
			in:   `{"idToken": "test token", "csrfToken": "csrf", "name": "Job Bohnson"}`,
			want: &LoginRequest{
				Name:      "Job Bohnson",
				IDToken:   "test token",
				CSRFToken: "csrf",
			},
		},
		{
			desc:    "missing id token",
			in:      `{"csrfToken": "csrf"}`,
			wantErr: true,
		},
		{
			desc:    "empty is malformed",
			in:      "",
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			got, err := parseLoginRequest(strings.NewReader(test.in))
			if test.wantErr {
				if err == nil {
					t.Fatal("no error was returned, but one was expected")
				}
				return
			}
			if err != nil {
				t.Fatalf("parseLoginRequest: %v", err)
			}

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Fatalf("unexpected login request (-want +got)\n%s", diff)
			}
		})
	}
}

package fireauth

import (
	"testing"
	"time"

	"github.com/Silicon-Ally/silicon-starter/authn"
	"github.com/google/go-cmp/cmp"

	firebaseauth "firebase.google.com/go/v4/auth"
)

func TestParseAuthTime(t *testing.T) {
	tests := []struct {
		desc    string
		in      map[string]interface{}
		want    time.Time
		wantErr bool
	}{
		{
			desc: "valid claims",
			in: map[string]interface{}{
				"auth_time": float64(123456789),
			},
			want: time.Unix(123456789, 0),
		},
		{
			desc: "invalid 'auth_time' type",
			in: map[string]interface{}{
				"auth_time": 123456789,
			},
			wantErr: true,
		},
		{
			desc: "no 'auth_time' claim",
			in: map[string]interface{}{
				"some_other_claim": "hello",
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			got, err := parseAuthTime(test.in)
			if test.wantErr {
				if err == nil {
					t.Fatal("no error was returned, but one was expected")
				}
				return
			}
			if err != nil {
				t.Fatalf("parseAuthTime: %v", err)
			}

			if !got.Equal(test.want) {
				t.Errorf("parseAuthTime = %q, want %q", got, test.want)
			}
		})
	}
}

func TestUserInfoFromToken(t *testing.T) {
	tests := []struct {
		desc    string
		in      *firebaseauth.Token
		want    *authn.UserInfo
		wantErr bool
	}{
		{
			desc: "valid token",
			in: &firebaseauth.Token{
				UID: "user-id",
				Firebase: firebaseauth.FirebaseInfo{
					SignInProvider: "google.com",
					Identities: map[string]interface{}{
						"email": []interface{}{
							"test-email@example.com",
						},
					},
				},
			},
			want: &authn.UserInfo{
				UserID:       "user-id",
				AuthProvider: authn.Google,
				Email:        "test-email@example.com",
			},
		},
		{
			desc: "invalid provider",
			in: &firebaseauth.Token{
				UID: "user-id",
				Firebase: firebaseauth.FirebaseInfo{
					SignInProvider: "unknown-provider-biz",
					Identities: map[string]interface{}{
						"email": []interface{}{
							"test-email@example.com",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			desc: "no email",
			in: &firebaseauth.Token{
				UID: "user-id",
				Firebase: firebaseauth.FirebaseInfo{
					SignInProvider: "google.com",
					Identities:     map[string]interface{}{},
				},
			},
			wantErr: true,
		},
		{
			desc: "bad email type",
			in: &firebaseauth.Token{
				UID: "user-id",
				Firebase: firebaseauth.FirebaseInfo{
					SignInProvider: "google.com",
					Identities: map[string]interface{}{
						// While this seems reasonable, it isn't how Firebase packages
						// identities.
						"email": "test-email@example.com",
					},
				},
			},
			wantErr: true,
		},
		{
			desc: "too many emails",
			in: &firebaseauth.Token{
				UID: "user-id",
				Firebase: firebaseauth.FirebaseInfo{
					SignInProvider: "google.com",
					Identities: map[string]interface{}{
						"email": []interface{}{
							"test-email@example.com",
							"another-test-email@example.com",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			desc: "bad email sub-type",
			in: &firebaseauth.Token{
				UID: "user-id",
				Firebase: firebaseauth.FirebaseInfo{
					SignInProvider: "google.com",
					Identities: map[string]interface{}{
						"email": []interface{}{
							123, // This is definitely not an email
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			got, err := userInfoFromToken(test.in)
			if test.wantErr {
				if err == nil {
					t.Fatal("no error was returned, but one was expected")
				}
				return
			}
			if err != nil {
				t.Fatalf("userInfoFromToken: %v", err)
			}

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("unexpected firebase token (-want +got)\n%s", diff)
			}
		})
	}
}

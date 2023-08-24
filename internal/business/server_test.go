package business

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"pg-to-es/internal/contract"
	"pg-to-es/internal/mock"
	"pg-to-es/internal/model"
	"reflect"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	type args struct {
		es      contract.Elastic
		port    int
		esIndex string
	}
	tests := []struct {
		name string
		args args
		want *Server
	}{
		{
			name: "should generate a not nil server instance, with exact provided instance and ReadTimeout & WriteTimeout setb to 5 seconds",
			args: args{
				es:      nil,
				port:    0,
				esIndex: "",
			},
			want: &Server{
				srv: &http.Server{
					Addr:         fmt.Sprintf(":%d", 0),
					ReadTimeout:  5 * time.Second,
					WriteTimeout: 5 * time.Second,
					Handler:      nil,
				},
				es:      nil,
				esIndex: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewServer(tt.args.es, tt.args.port, tt.args.esIndex); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewServer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServer_InitRoutes(t *testing.T) {
	type fields struct {
		srv     *http.Server
		es      contract.Elastic
		esIndex string
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "handler must get initialized",
			fields: fields{
				srv:     NewServer(nil, 0, "").srv,
				es:      nil,
				esIndex: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				srv:     tt.fields.srv,
				es:      tt.fields.es,
				esIndex: tt.fields.esIndex,
			}
			s.InitRoutes()
			assert.NotNil(t, s.srv.Handler, "handler is nil")
		})
	}
}

func TestServer_Start(t *testing.T) {
	type fields struct {
		srv     *http.Server
		es      contract.Elastic
		esIndex string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "nil handler should return error",
			fields: fields{
				srv:     NewServer(nil, 0, "").srv,
				es:      nil,
				esIndex: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				srv:     tt.fields.srv,
				es:      tt.fields.es,
				esIndex: tt.fields.esIndex,
			}
			if err := s.Start(); (err != nil) != tt.wantErr {
				t.Errorf("Server.Start() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestServer_Root(t *testing.T) {
	esMock := mock.NewElastic([]model.User{})
	type fields struct {
		srv     *http.Server
		es      contract.Elastic
		esIndex string
	}
	type args struct {
		method string
		target string
		body   io.Reader
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		returnStatus int
	}{
		{
			name: "should always return 200 OK",
			fields: fields{
				srv:     NewServer(esMock, 0, "").srv,
				es:      esMock,
				esIndex: "",
			},
			args: args{
				method: http.MethodGet,
				target: "/",
				body:   nil,
			},
			returnStatus: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				srv:     tt.fields.srv,
				es:      tt.fields.es,
				esIndex: tt.fields.esIndex,
			}
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tt.args.method, tt.args.target, tt.args.body)
			s.Root(w, req)
			assert.Equal(t, tt.returnStatus, w.Code, "status code must match")
		})
	}
}

func TestServer_SearchProjectsByUser(t *testing.T) {
	now := time.Now().Format(time.RFC3339)
	esMock := mock.NewElastic([]model.User{
		model.User{
			ID:        1,
			Name:      "Test user",
			CreatedAt: now,
			Projects: []model.Project{
				model.Project{
					ID:          1,
					Name:        "Test project",
					Slug:        "Test project slug",
					Description: "Test project description",
					CreatedAt:   now,
					Hashtags: []model.Hashtag{
						model.Hashtag{
							ID:        1,
							Name:      "TestHashTag",
							CreatedAt: now,
						},
					},
				},
			},
		},
	})
	server := NewServer(esMock, 0, "")
	server.InitRoutes()
	type fields struct {
		srv     *http.Server
		es      contract.Elastic
		esIndex string
	}
	type args struct {
		method string
		target string
		body   io.Reader
		vars   map[string]string
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		returnStatus int
	}{
		{
			name: "should return 400 BadRequest for invalid user id",
			fields: fields{
				srv:     server.srv,
				es:      esMock,
				esIndex: "",
			},
			args: args{
				method: http.MethodGet,
				target: "/search/user",
				body:   nil,
				vars: map[string]string{
					"userID": "abc",
				},
			},
			returnStatus: http.StatusBadRequest,
		},
		{
			name: "should return 200 OK for user present in engine",
			fields: fields{
				srv:     server.srv,
				es:      esMock,
				esIndex: "",
			},
			args: args{
				method: http.MethodGet,
				target: "/search/user",
				body:   nil,
				vars: map[string]string{
					"userID": "1",
				},
			},
			returnStatus: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tt.args.method, tt.args.target, tt.args.body)
			req = mux.SetURLVars(req, tt.args.vars)
			server.SearchProjectsByUser(w, req)
			assert.Equal(t, tt.returnStatus, w.Code, "status code must match")
		})
	}
}

func TestServer_SearchProjectsByHashtag(t *testing.T) {
	now := time.Now().Format(time.RFC3339)
	esMock := mock.NewElastic([]model.User{
		model.User{
			ID:        1,
			Name:      "Test user",
			CreatedAt: now,
			Projects: []model.Project{
				model.Project{
					ID:          1,
					Name:        "Test project",
					Slug:        "Test project slug",
					Description: "Test project description",
					CreatedAt:   now,
					Hashtags: []model.Hashtag{
						model.Hashtag{
							ID:        1,
							Name:      "TestHashTag",
							CreatedAt: now,
						},
					},
				},
			},
		},
	})
	server := NewServer(esMock, 0, "")
	server.InitRoutes()
	type fields struct {
		srv     *http.Server
		es      contract.Elastic
		esIndex string
	}
	type args struct {
		method string
		target string
		body   io.Reader
		vars   map[string]string
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		returnStatus int
	}{
		{
			name: "should return 200 OK for user present in engine",
			fields: fields{
				srv:     server.srv,
				es:      esMock,
				esIndex: "",
			},
			args: args{
				method: http.MethodGet,
				target: "/search/hashtags",
				body:   nil,
				vars: map[string]string{
					"hashtag": "1",
				},
			},
			returnStatus: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tt.args.method, tt.args.target, tt.args.body)
			req = mux.SetURLVars(req, tt.args.vars)
			server.SearchProjectsByHashtag(w, req)
			assert.Equal(t, tt.returnStatus, w.Code, "status code must match")
		})
	}
}

func TestServer_FuzzySearchProjects(t *testing.T) {
	now := time.Now().Format(time.RFC3339)
	esMock := mock.NewElastic([]model.User{
		model.User{
			ID:        1,
			Name:      "Test user",
			CreatedAt: now,
			Projects: []model.Project{
				model.Project{
					ID:          1,
					Name:        "Test project",
					Slug:        "Test project slug",
					Description: "Test project description",
					CreatedAt:   now,
					Hashtags: []model.Hashtag{
						model.Hashtag{
							ID:        1,
							Name:      "TestHashTag",
							CreatedAt: now,
						},
					},
				},
			},
		},
	})
	server := NewServer(esMock, 0, "")
	server.InitRoutes()
	type fields struct {
		srv     *http.Server
		es      contract.Elastic
		esIndex string
	}
	type args struct {
		method string
		target string
		body   io.Reader
		vars   map[string]string
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		returnStatus int
	}{
		{
			name: "should return 200 OK for user present in engine",
			fields: fields{
				srv:     server.srv,
				es:      esMock,
				esIndex: "",
			},
			args: args{
				method: http.MethodGet,
				target: "/search/fuzzy",
				body:   nil,
				vars: map[string]string{
					"query": "Test",
				},
			},
			returnStatus: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tt.args.method, tt.args.target, tt.args.body)
			req = mux.SetURLVars(req, tt.args.vars)
			server.FuzzySearchProjects(w, req)
			assert.Equal(t, tt.returnStatus, w.Code, "status code must match")
		})
	}
}

func TestServer_encode(t *testing.T) {
	rr := httptest.NewRecorder()
	encode(rr, struct{}{})
	if rr.Result().StatusCode != 200 {
		t.Errorf("Status code returned, %d, did not match expected code %d", rr.Result().StatusCode, 200)
	}
	if rr.Result().Header.Get("Content-Type") != "application/json" {
		t.Errorf("Header value for `tracing-id`, %s, did not match expected value %s", rr.Result().Header.Get("Content-Type"), "application/json")
	}
}

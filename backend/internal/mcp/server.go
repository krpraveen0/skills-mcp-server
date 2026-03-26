package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/krpraveen0/skills-mcp-server/internal/cache"
	"github.com/krpraveen0/skills-mcp-server/internal/db"
	"github.com/krpraveen0/skills-mcp-server/pkg/models"
)

const protocolVersion = "2024-11-05"
const serverVersion   = "1.0.0"

// Server handles MCP JSON-RPC requests over HTTP.
type Server struct {
	db    *db.DB
	cache *cache.Redis
	cacheTTLSearch   time.Duration
	cacheTTLTrending time.Duration
	cacheTTLSkill    time.Duration
}

// NewServer creates a new MCP server handler.
func NewServer(database *db.DB, redisCache *cache.Redis,
	ttlSearch, ttlTrending, ttlSkill int) *Server {
	return &Server{
		db:               database,
		cache:            redisCache,
		cacheTTLSearch:   time.Duration(ttlSearch) * time.Second,
		cacheTTLTrending: time.Duration(ttlTrending) * time.Second,
		cacheTTLSkill:    time.Duration(ttlSkill) * time.Second,
	}
}

// Handle is the Gin handler for POST /mcp
func (s *Server) Handle(c *gin.Context) {
	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, Response{
			JSONRPC: "2.0",
			Error:   &RPCError{Code: ErrParse, Message: "Parse error: " + err.Error()},
		})
		return
	}

	if req.JSONRPC != "2.0" {
		c.JSON(http.StatusOK, Response{
			JSONRPC: "2.0", ID: req.ID,
			Error: &RPCError{Code: ErrInvalidRequest, Message: "Invalid JSON-RPC version"},
		})
		return
	}

	resp := s.dispatch(c.Request.Context(), &req)
	c.JSON(http.StatusOK, resp)
}

// dispatch routes a JSON-RPC method to the correct handler.
func (s *Server) dispatch(ctx context.Context, req *Request) Response {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolsCall(ctx, req)
	default:
		return Response{
			JSONRPC: "2.0", ID: req.ID,
			Error: &RPCError{Code: ErrMethodNotFound, Message: "Method not found: " + req.Method},
		}
	}
}

// handleInitialize responds to the MCP initialize handshake.
func (s *Server) handleInitialize(req *Request) Response {
	return Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: InitializeResult{
			ProtocolVersion: protocolVersion,
			Capabilities: Capabilities{
				Tools: &ToolsCapability{ListChanged: false},
			},
			ServerInfo: ServerInfo{
				Name:    "skills-mcp-server",
				Version: serverVersion,
			},
		},
	}
}

// handleToolsList returns all available tool definitions.
func (s *Server) handleToolsList(req *Request) Response {
	return Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  ToolsListResult{Tools: toolDefinitions()},
	}
}

// handleToolsCall dispatches a tool call to the correct handler.
func (s *Server) handleToolsCall(ctx context.Context, req *Request) Response {
	// Re-marshal params to get a typed CallToolParams
	paramsBytes, err := json.Marshal(req.Params)
	if err != nil {
		return rpcError(req.ID, ErrInvalidParams, "invalid params")
	}
	var params CallToolParams
	if err := json.Unmarshal(paramsBytes, &params); err != nil {
		return rpcError(req.ID, ErrInvalidParams, "invalid params: "+err.Error())
	}

	var result *CallToolResult
	switch params.Name {
	case "search_skills":
		result, err = s.toolSearchSkills(ctx, params.Arguments)
	case "get_skill_detail":
		result, err = s.toolGetSkillDetail(ctx, params.Arguments)
	case "list_trending_skills":
		result, err = s.toolListTrendingSkills(ctx, params.Arguments)
	case "submit_skill":
		result, err = s.toolSubmitSkill(ctx, params.Arguments)
	default:
		return rpcError(req.ID, ErrMethodNotFound, "unknown tool: "+params.Name)
	}

	if err != nil {
		return Response{
			JSONRPC: "2.0", ID: req.ID,
			Result: &CallToolResult{
				IsError: true,
				Content: []ContentBlock{{Type: "text", Text: err.Error()}},
			},
		}
	}

	return Response{JSONRPC: "2.0", ID: req.ID, Result: result}
}

// --- Tool handlers ---

func (s *Server) toolSearchSkills(ctx context.Context, args map[string]any) (*CallToolResult, error) {
	query := stringArg(args, "query", "")
	limit := intArg(args, "limit", 10)
	offset := intArg(args, "offset", 0)

	cacheKey := fmt.Sprintf("mcp:search:%s:%d:%d", query, limit, offset)

	var resp models.SearchResponse
	if err := s.cache.Get(ctx, cacheKey, &resp); err != nil {
		// Cache miss
		skills, total, err := s.db.SearchSkills(ctx, query, nil, limit, offset)
		if err != nil {
			return nil, fmt.Errorf("search failed: %w", err)
		}
		resp = models.SearchResponse{Skills: skills, Total: total, Limit: limit, Offset: offset}
		s.cache.Set(ctx, cacheKey, resp, s.cacheTTLSearch)
	}

	// Strip full content from search results for brevity
	for i := range resp.Skills {
		resp.Skills[i].Content = ""
	}

	out, _ := json.MarshalIndent(resp, "", "  ")
	return &CallToolResult{
		Content: []ContentBlock{{Type: "text", Text: string(out)}},
	}, nil
}

func (s *Server) toolGetSkillDetail(ctx context.Context, args map[string]any) (*CallToolResult, error) {
	id := stringArg(args, "id", "")
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}

	cacheKey := "mcp:skill:" + id
	var skill models.Skill
	if err := s.cache.Get(ctx, cacheKey, &skill); err != nil {
		s2, err := s.db.GetSkillByID(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("skill not found: %s", id)
		}
		skill = *s2
		s.cache.Set(ctx, cacheKey, skill, s.cacheTTLSkill)
	}

	out, _ := json.MarshalIndent(skill, "", "  ")
	return &CallToolResult{
		Content: []ContentBlock{{Type: "text", Text: string(out)}},
	}, nil
}

func (s *Server) toolListTrendingSkills(ctx context.Context, args map[string]any) (*CallToolResult, error) {
	limit    := intArg(args, "limit", 20)
	category := stringArg(args, "category", "")

	cacheKey := fmt.Sprintf("mcp:trending:%d:%s", limit, category)
	var skills []models.Skill
	if err := s.cache.Get(ctx, cacheKey, &skills); err != nil {
		var err2 error
		skills, err2 = s.db.ListTrendingSkills(ctx, limit, category)
		if err2 != nil {
			return nil, fmt.Errorf("fetch trending failed: %w", err2)
		}
		// Strip content for trending list
		for i := range skills {
			skills[i].Content = ""
		}
		s.cache.Set(ctx, cacheKey, skills, s.cacheTTLTrending)
	}

	out, _ := json.MarshalIndent(skills, "", "  ")
	return &CallToolResult{
		Content: []ContentBlock{{Type: "text", Text: string(out)}},
	}, nil
}

func (s *Server) toolSubmitSkill(ctx context.Context, args map[string]any) (*CallToolResult, error) {
	githubURL := stringArg(args, "github_url", "")
	if githubURL == "" {
		return nil, fmt.Errorf("github_url is required")
	}
	notes := stringArg(args, "notes", "")

	sub := &models.SkillSubmission{
		ID:          uuid.New().String(),
		GitHubURL:   githubURL,
		SubmittedBy: "mcp",
		SubmittedAt: time.Now(),
		Status:      "pending",
		Notes:       notes,
	}
	if err := s.db.CreateSubmission(ctx, sub); err != nil {
		return nil, fmt.Errorf("failed to record submission: %w", err)
	}

	return &CallToolResult{
		Content: []ContentBlock{{
			Type: "text",
			Text: fmt.Sprintf(`{"status":"queued","id":"%s","message":"Your skill has been queued for indexing. It will appear in search results after the next crawl."}`, sub.ID),
		}},
	}, nil
}

// --- Helpers ---

func rpcError(id any, code int, msg string) Response {
	return Response{
		JSONRPC: "2.0", ID: id,
		Error: &RPCError{Code: code, Message: msg},
	}
}

func stringArg(args map[string]any, key, def string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return def
}

func intArg(args map[string]any, key string, def int) int {
	if v, ok := args[key]; ok {
		switch n := v.(type) {
		case float64:
			return int(n)
		case int:
			return n
		}
	}
	return def
}

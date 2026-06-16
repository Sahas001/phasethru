package agents

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

type Orchestrator struct {
	GitHub *GitHubClient
	// LLMClient could be an interface to the Gemini API
}

func NewOrchestrator(gh *GitHubClient) *Orchestrator {
	return &Orchestrator{
		GitHub: gh,
	}
}

// RunWeeklyRoutine executes the full multi-agent pipeline
func (o *Orchestrator) RunWeeklyRoutine(ctx context.Context, rootDir string) error {
	fmt.Println("🚀 Starting Weekly Agent Routine...")

	chunks, err := o.chunkCodebase(rootDir)
	if err != nil {
		return fmt.Errorf("failed to chunk codebase: %w", err)
	}

	fmt.Printf("📦 Divided codebase into %d chunks.\n", len(chunks))

	var prIDs []int

	// 1. Run Reviewer & Suggester Agents
	for i := range chunks {
		fmt.Printf("🕵️  Agent assigned to chunk %d...\n", i)
		
		// In a real implementation, we would send 'chunk' + ReviewerPrompt to the LLM.
		// Here we mock the LLM suggesting a change.
		branchName := fmt.Sprintf("agent-suggestion-chunk-%d", i)
		err := o.GitHub.CreateBranch(ctx, branchName)
		if err != nil {
			continue
		}

		// Mock file change
		_ = o.GitHub.CommitFile(ctx, branchName, "mock_file.go", "// Improvements", "feat: agent improvements")
		
		prID, err := o.GitHub.CreatePullRequest(ctx, "Agent Suggestion: Code Quality", "Automated improvements.", branchName, "main")
		if err == nil {
			prIDs = append(prIDs, prID)
		}
	}

	// 2. Run Website Manager Agent
	fmt.Println("🌐 Website Manager Agent reviewing 'website/'...")
	webBranch := "agent-website-update"
	_ = o.GitHub.CreateBranch(ctx, webBranch)
	_ = o.GitHub.CommitFile(ctx, webBranch, "website/index.html", "<!-- SEO Improvements -->", "chore: update website")
	webPR, err := o.GitHub.CreatePullRequest(ctx, "Agent Suggestion: Website Update", "Automated website maintenance.", webBranch, "main")
	if err == nil {
		prIDs = append(prIDs, webPR)
	}

	// 3. Summon the Judge
	if len(prIDs) > 0 {
		fmt.Printf("⚖️  Judge evaluating %d pull requests...\n", len(prIDs))
		// In a real implementation, the JudgePrompt would be used alongside PR diffs.
		bestPR := prIDs[0] // Mocking the judge picking the first one
		
		comment := "👑 **Judge Agent:** This PR has been selected as the most impactful change of the week. Great job, Agent!"
		err = o.GitHub.PostComment(ctx, bestPR, comment)
		if err != nil {
			return fmt.Errorf("failed to post judge comment: %w", err)
		}
	}

	fmt.Println("✅ Weekly Agent Routine completed successfully.")
	return nil
}

// chunkCodebase reads all .go files and groups them into logical chunks
func (o *Orchestrator) chunkCodebase(rootDir string) ([][]string, error) {
	var files []string
	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".go") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Simple chunking: 5 files per chunk
	var chunks [][]string
	chunkSize := 5
	for i := 0; i < len(files); i += chunkSize {
		end := i + chunkSize
		if end > len(files) {
			end = len(files)
		}
		chunks = append(chunks, files[i:end])
	}

	return chunks, nil
}

package graveyard

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// GraveyardFilename is the name of the graveyard file
	GraveyardFilename = "GRAVEYARD.md"

	// GraveyardHeader is the header content for a new GRAVEYARD.md file
	GraveyardHeader = `# Graveyard

Your fallen pets rest here. Each death is a reminder to balance work and rest.

## Deaths

| Date | Time | Overtime | Cause |
|------|------|----------|-------|
`

	// ResurrectionsSectionHeader is the header for the resurrections section
	ResurrectionsSectionHeader = `
## Resurrections

| Date | Time | Notes |
|------|------|-------|
`
)

// Graveyard manages the death records and GRAVEYARD.md file
type Graveyard struct {
	// DataDir is the directory where GRAVEYARD.md is stored
	DataDir string

	// Records contains all loaded death records
	Records []DeathRecord

	// Resurrections contains all loaded resurrection records
	Resurrections []ResurrectionRecord

	mu sync.RWMutex
}

// NewGraveyard creates a new Graveyard instance
func NewGraveyard(dataDir string) *Graveyard {
	return &Graveyard{
		DataDir:       dataDir,
		Records:       make([]DeathRecord, 0),
		Resurrections: make([]ResurrectionRecord, 0),
	}
}

// filePath returns the full path to GRAVEYARD.md
func (g *Graveyard) filePath() string {
	return filepath.Join(g.DataDir, GraveyardFilename)
}

// LogDeath appends a death record to GRAVEYARD.md
func (g *Graveyard) LogDeath(record DeathRecord) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Ensure data directory exists
	if err := os.MkdirAll(g.DataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	path := g.filePath()

	// Check if file exists
	_, err := os.Stat(path)
	fileExists := !os.IsNotExist(err)

	// Open file for appending (create if not exists)
	var file *os.File
	if fileExists {
		file, err = os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	} else {
		file, err = os.Create(path)
		if err != nil {
			return fmt.Errorf("failed to create graveyard file: %w", err)
		}
		// Write header for new file
		if _, err := file.WriteString(GraveyardHeader); err != nil {
			file.Close()
			return fmt.Errorf("failed to write header: %w", err)
		}
	}
	if err != nil {
		return fmt.Errorf("failed to open graveyard file: %w", err)
	}
	defer file.Close()

	// Append the record
	row := record.ToMarkdownRow()
	if _, err := file.WriteString(row + "\n"); err != nil {
		return fmt.Errorf("failed to write death record: %w", err)
	}

	// Add to in-memory records
	g.Records = append(g.Records, record)

	return nil
}

// LoadRecords parses existing GRAVEYARD.md and loads records into memory
func (g *Graveyard) LoadRecords() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	path := g.filePath()

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// File doesn't exist, that's OK - no records yet
		g.Records = make([]DeathRecord, 0)
		g.Resurrections = make([]ResurrectionRecord, 0)
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open graveyard file: %w", err)
	}
	defer file.Close()

	g.Records = make([]DeathRecord, 0)
	g.Resurrections = make([]ResurrectionRecord, 0)

	// Parse the markdown table for deaths
	// Format: | Date | Time | Overtime | Cause |
	// Example: | 2026-01-21 | 14:32 | 5m 23s | overtime |
	deathRowRegex := regexp.MustCompile(`^\|\s*(\d{4}-\d{2}-\d{2})\s*\|\s*(\d{2}:\d{2})\s*\|\s*([^|]+)\s*\|\s*([^|]+)\s*\|$`)

	// Parse the markdown table for resurrections
	// Format: | Date | Time | Notes |
	// Example: | 2026-01-21 | 14:45 | Recovered from death #1 |
	resurrectionRowRegex := regexp.MustCompile(`^\|\s*(\d{4}-\d{2}-\d{2})\s*\|\s*(\d{2}:\d{2})\s*\|\s*Recovered from death #(\d+)\s*\|$`)

	inResurrectionSection := false
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Check for section markers
		if strings.HasPrefix(line, "## Resurrections") {
			inResurrectionSection = true
			continue
		}
		if strings.HasPrefix(line, "## Deaths") {
			inResurrectionSection = false
			continue
		}

		// Try resurrection row first (more specific pattern)
		if resMatches := resurrectionRowRegex.FindStringSubmatch(line); resMatches != nil {
			dateStr := resMatches[1]
			timeStr := resMatches[2]
			datetimeStr := dateStr + " " + timeStr
			timestamp, err := time.Parse("2006-01-02 15:04", datetimeStr)
			if err != nil {
				continue // Skip invalid rows
			}

			deathIndex, err := strconv.Atoi(resMatches[3])
			if err != nil {
				continue // Skip invalid rows
			}

			record := ResurrectionRecord{
				Timestamp:  timestamp,
				DeathIndex: deathIndex,
			}
			g.Resurrections = append(g.Resurrections, record)
			continue
		}

		// Try death row (if not in resurrection section)
		if !inResurrectionSection {
			if deathMatches := deathRowRegex.FindStringSubmatch(line); deathMatches != nil {
				dateStr := deathMatches[1]
				timeStr := deathMatches[2]
				datetimeStr := dateStr + " " + timeStr
				timestamp, err := time.Parse("2006-01-02 15:04", datetimeStr)
				if err != nil {
					continue // Skip invalid rows
				}

				overtimeStr := strings.TrimSpace(deathMatches[3])
				overtimeSeconds := parseDuration(overtimeStr)
				cause := strings.TrimSpace(deathMatches[4])

				record := DeathRecord{
					Timestamp:       timestamp,
					OvertimeSeconds: overtimeSeconds,
					SessionEarned:   0, // Not stored in file, unknown for historical records
					Cause:           cause,
				}
				g.Records = append(g.Records, record)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading graveyard file: %w", err)
	}

	return nil
}

// GetDeathCount returns the number of recorded deaths
func (g *Graveyard) GetDeathCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.Records)
}

// GetResurrectionCount returns the number of recorded resurrections
func (g *Graveyard) GetResurrectionCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.Resurrections)
}

// IsDead returns true if the pet is currently dead (last event was death, not resurrection)
func (g *Graveyard) IsDead() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	// Pet is dead if there are more deaths than resurrections
	return len(g.Records) > len(g.Resurrections)
}

// LogResurrection appends a resurrection record to GRAVEYARD.md
func (g *Graveyard) LogResurrection() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Calculate death index (which death we're recovering from)
	deathIndex := len(g.Records)
	if deathIndex == 0 {
		return fmt.Errorf("cannot resurrect: no recorded deaths")
	}

	// Check if already resurrected from the most recent death
	if len(g.Resurrections) >= len(g.Records) {
		return fmt.Errorf("cannot resurrect: pet is not dead")
	}

	record := ResurrectionRecord{
		Timestamp:  time.Now(),
		DeathIndex: deathIndex,
	}

	// Ensure data directory exists
	if err := os.MkdirAll(g.DataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	path := g.filePath()

	// Read existing file content
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read graveyard file: %w", err)
	}

	contentStr := string(content)

	// Check if Resurrections section exists
	if !strings.Contains(contentStr, "## Resurrections") {
		// Append the resurrections section header
		contentStr += ResurrectionsSectionHeader
	}

	// Append the resurrection record
	row := record.ToMarkdownRow()
	contentStr += row + "\n"

	// Write back to file
	if err := os.WriteFile(path, []byte(contentStr), 0644); err != nil {
		return fmt.Errorf("failed to write resurrection record: %w", err)
	}

	// Add to in-memory records
	g.Resurrections = append(g.Resurrections, record)

	return nil
}

// GetStreakInfo calculates and returns streak information from death records
func (g *Graveyard) GetStreakInfo() StreakInfo {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return CalculateStreak(g.Records)
}

// parseDuration parses a duration string like "5m 23s" or "5m" or "30s" into seconds
func parseDuration(s string) int {
	s = strings.TrimSpace(s)
	totalSeconds := 0

	// Match minutes
	minuteRegex := regexp.MustCompile(`(\d+)m`)
	if matches := minuteRegex.FindStringSubmatch(s); matches != nil {
		if mins, err := strconv.Atoi(matches[1]); err == nil {
			totalSeconds += mins * 60
		}
	}

	// Match seconds
	secondRegex := regexp.MustCompile(`(\d+)s`)
	if matches := secondRegex.FindStringSubmatch(s); matches != nil {
		if secs, err := strconv.Atoi(matches[1]); err == nil {
			totalSeconds += secs
		}
	}

	return totalSeconds
}

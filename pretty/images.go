package pretty

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

// PrintImages displays Docker images in a pretty format
func PrintImages(args []string) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating Docker client: %v\n", err)
		os.Exit(1)
	}
	defer cli.Close()

	ctx := context.Background()

	images, err := cli.ImageList(ctx, image.ListOptions{All: false})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing images: %v\n", err)
		os.Exit(1)
	}

	if len(images) == 0 {
		gray.Println("No images found")
		return
	}

	// Print header
	cyan.Println("\n╭─────────────────────────────────────────────────────────────────────────────────╮")
	cyan.Printf("│ %-78s │\n", "IMAGES")
	cyan.Println("├─────────────────────────────────────────────────────────────────────────────────┤")

	var totalSize int64

	// Print images
	for _, img := range images {
		// Get repository and tag
		repoTag := "<none>:<none>"
		if len(img.RepoTags) > 0 {
			repoTag = img.RepoTags[0]
		}

		// Truncate if too long
		if len(repoTag) > 45 {
			repoTag = repoTag[:42] + "..."
		}

		// Format size
		size := formatSize(img.Size)

		// Format created time
		created := formatCreatedTime(img.Created)

		// Image ID (short)
		imageID := img.ID
		if strings.HasPrefix(imageID, "sha256:") {
			imageID = imageID[7:19] // Get first 12 chars after sha256:
		}

		cyan.Print("│ ")
		blue.Printf("%-45s", repoTag)
		fmt.Printf(" │ ")
		green.Printf("%-10s", size)
		fmt.Printf(" │ ")
		gray.Printf("%-12s", imageID)
		cyan.Println(" │")

		cyan.Print("│   ")
		gray.Printf("⏱ Created: %s", created)
		fmt.Printf("%*s", 78-len(created)-14, "")
		cyan.Println("│")
		cyan.Println("├─────────────────────────────────────────────────────────────────────────────────┤")

		totalSize += img.Size
	}

	cyan.Println("╰─────────────────────────────────────────────────────────────────────────────────╯\n")

	// Summary
	fmt.Printf("Total: %d images", len(images))
	if totalSize > 0 {
		green.Printf(" (Total size: %s)", formatSize(totalSize))
	}
	fmt.Println()
}

func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

func formatCreatedTime(timestamp int64) string {
	created := time.Unix(timestamp, 0)
	duration := time.Since(created)

	switch {
	case duration < time.Minute:
		return fmt.Sprintf("%d seconds ago", int(duration.Seconds()))
	case duration < time.Hour:
		return fmt.Sprintf("%d minutes ago", int(duration.Minutes()))
	case duration < 24*time.Hour:
		return fmt.Sprintf("%d hours ago", int(duration.Hours()))
	case duration < 7*24*time.Hour:
		return fmt.Sprintf("%d days ago", int(duration.Hours()/24))
	case duration < 30*24*time.Hour:
		return fmt.Sprintf("%d weeks ago", int(duration.Hours()/(24*7)))
	case duration < 365*24*time.Hour:
		return fmt.Sprintf("%d months ago", int(duration.Hours()/(24*30)))
	default:
		return fmt.Sprintf("%d years ago", int(duration.Hours()/(24*365)))
	}
}

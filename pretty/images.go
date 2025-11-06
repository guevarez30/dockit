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
	fmt.Println()
	cyan.Println("IMAGES")
	cyan.Println(strings.Repeat("─", 90))

	var totalSize int64

	// Print images
	for _, img := range images {
		// Image ID (short)
		imageID := img.ID
		if strings.HasPrefix(imageID, "sha256:") {
			imageID = imageID[7:19] // Get first 12 chars after sha256:
		}
		idWidth := 12
		idPadded := imageID + strings.Repeat(" ", idWidth-len(imageID))

		// Get repository and tag
		repoTag := "<none>:<none>"
		if len(img.RepoTags) > 0 {
			repoTag = img.RepoTags[0]
		}
		repoWidth := 40
		if len(repoTag) > repoWidth {
			repoTag = repoTag[:repoWidth-3] + "..."
		}
		repoPadded := repoTag + strings.Repeat(" ", repoWidth-len(repoTag))

		// Format size
		size := formatSize(img.Size)
		sizeWidth := 12
		sizePadded := size + strings.Repeat(" ", sizeWidth-len(size))

		// Format created time
		created := formatCreatedTime(img.Created)

		// Print main line
		gray.Print(idPadded)
		gray.Print(" │ ")
		blue.Print(repoPadded)
		gray.Print(" │ ")
		green.Print(sizePadded)
		gray.Print("│ ")
		gray.Println(created)

		fmt.Println()
		totalSize += img.Size
	}

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

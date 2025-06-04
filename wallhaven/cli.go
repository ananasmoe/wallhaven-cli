package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

func search(images ImagesResponse) (string, error) {
	selections := []string{}

	if len(images.Images) == 0 {
		return "", fmt.Errorf("No collections from this user.")
	}

	if images.Meta.LastPage > page {
		selections = append(selections, "Next page -->")
	}

	if page > 1 {
		selections = append(selections, "Previous page <--")
	}

	for _, v := range images.Images {
		preview := fmt.Sprintf("%v (%v)", v.Resolution, v.ImageURL)
		selections = append(selections, preview)
	}

	selection, err := ShowSelection(selections, true)
	if err != nil {
		return "", err
	}

	return selection, nil
}

func Search(cmd *cobra.Command, args []string) error {
	var url string

	query := strings.Join(args, " ")

	for url == "" {
		images, err := SearchAPI(query, page)
		if err != nil {
			return err
		}

		selection, err := search(images)
		if err != nil {
			return err
		}

		if strings.Contains(selection, "-->") {
			page++
		} else if strings.Contains(selection, "<--") {
			page--
		} else {
			url = re.FindStringSubmatch(selection)[1]
		}
	}

	err := DLSave(url, nil)
	if err != nil {
		return err
	}

	fmt.Printf("[download] %v\n", url)
	return nil
}

func Download(cmd *cobra.Command, args []string) error {
	for _, v := range args {
		err := DownloadAPI(v)
		if err != nil {
			return err
		}

		fmt.Printf("[download] %s", v)
	}

	return nil
}

func Edit(cmd *cobra.Command, args []string) error {
	command := exec.Command(config.Editor, configFile)

	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	return command.Run()
}

func Collection(cmd *cobra.Command, args []string) error {
	username := args[0]
	selections := []string{}

	collections, err := CollectionsAPI(username)
	if err != nil {
		return err
	}

	if len(collections.Collections) == 0 {
		return fmt.Errorf("This user has no collection.")
	}

	for _, v := range collections.Collections {
		selection := fmt.Sprintf("%v (%v)", v.Label, v.ID)
		selections = append(selections, selection)
	}

	SelectedID, err := ShowSelection(selections, false)
	if err != nil {
		return err
	}

	id := re.FindStringSubmatch(SelectedID)[1]

	if all {
		for {
			images, err := CollectionAPI(username, id, page)
			if err != nil {
				return err
			}

			if len(images.Images) == 0 {
				return fmt.Errorf("This collection has no images.")
			}

			for _, v := range images.Images {
				err := DLSave(v.ImageURL, nil)
				if err != nil {
					return err
				}

				fmt.Printf("[download] %s\r\n", v.ImageID)
			}
			if images.Meta.LastPage > page {
				page++
			} else {
				return nil
			}
		}
	}

	var url string

	for url == "" {
		images, err := CollectionAPI(username, id, page)
		if err != nil {
			return err
		}

		for _, v := range images.Images {
			preview := fmt.Sprintf("%v (%v)", v.Resolution, v.ImageURL)
			selections = append(selections, preview)
		}

		selection, err := search(images)
		if err != nil {
			return err
		}

		if strings.Contains(selection, "-->") {
			page++
		} else if strings.Contains(selection, "<--") {
			page--
		} else {
			url = re.FindStringSubmatch(selection)[1]
		}
	}

	ImageURL := re.FindStringSubmatch(url)[0]

	err = DLSave(ImageURL, nil)
	if err != nil {
		return err
	}

	fmt.Printf("[download] %v", ImageURL)
	return nil
}

func ShowSelection(items []string, preview bool) (string, error) {
	input := strings.Join(items, "\n")

	cmd := exec.Command("fzf")
	if preview {
		cmd.Args = append(cmd.Args, "--preview=wallhaven preview {}")
	}

	cmd.Stdin = strings.NewReader(input)

	selected, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(selected)), nil
}

func Preview(cmd *cobra.Command, args []string) error {
	var command *exec.Cmd
	var file string

	text := args[0]

	if strings.Contains(text, "-->") || strings.Contains(text, "<--") {
		return nil
	}

	fzf_preview_lines := os.Getenv("FZF_PREVIEW_LINES")
	fzf_preview_columns := os.Getenv("FZF_PREVIEW_COLUMNS")

	// Check for terminal capabilities
	kittySupport := os.Getenv("KITTY_WINDOW_ID") != ""
	sixelSupport := checkSixelSupport()

	url := re.FindStringSubmatch(text)[1]

	// Priority: Kitty > Sixel > Chafa
	switch {
	case kittySupport:
		command = exec.Command(
			"kitty",
			"icat",
			"--clear",
			"--transfer-mode=memory",
			"--unicode-placeholder",
			"--stdin=no",
			fmt.Sprintf("--place=%sx%s@0x0", fzf_preview_columns, fzf_preview_lines),
			"--scale-up",
			url,
		)

	case sixelSupport:
		// Download image for sixel conversion
		err := DLSave(url, &config.TempFolder)
		if err != nil {
			return err
		}

		// Get terminal dimensions
		width, height, err := getPreviewPaneDimensions()
		if err != nil {
			// Fallback to FZF environment variables if available
			if fzfLines := os.Getenv("FZF_PREVIEW_LINES"); fzfLines != "" {
				if fzfCols := os.Getenv("FZF_PREVIEW_COLUMNS"); fzfCols != "" {
					width = fzfCols
					height = fzfLines
				}
			} else {
				// Last resort fallback
				width = "80"
				height = "24"
			}
		}

		SplitUrl := strings.Split(url, "/")
		FileName := SplitUrl[len(SplitUrl)-1]
		file = filepath.Join(config.TempFolder, FileName)

		// Convert terminal cells to pixels for sixel
		pixelWidth, pixelHeight := cellsToPixels(width, height)

		// Try img2sixel first (better quality)
		if path, err := exec.LookPath("img2sixel"); err == nil {
			command = exec.Command(
				path,
				"-w", pixelWidth,
				file,
			)
		} else if path, err := exec.LookPath("magick"); err == nil {
			// Fallback to ImageMagick convert
			command = exec.Command(
				path,
				file,
				"-resize", fmt.Sprintf("%sx%s>", pixelWidth, pixelHeight),
				"-quality", "100",
				"sixel:-",
			)
		} else {
			return fmt.Errorf("sixel support detected but no converter found (install libsixel or imagemagick)")
		}

	default:
		// Fallback to chafa
		path, err := exec.LookPath("chafa")
		if err != nil {
			return fmt.Errorf("no supported image preview method found")
		}

		err = DLSave(url, &config.TempFolder)
		if err != nil {
			return err
		}

		SplitUrl := strings.Split(url, "/")
		FileName := SplitUrl[len(SplitUrl)-1]
		file = filepath.Join(config.TempFolder, FileName)

		command = exec.Command(
			path,
			file,
			fmt.Sprintf("--size=%sx%s", fzf_preview_columns, fzf_preview_lines),
			"--clear",
		)
	}

	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	if err := command.Run(); err != nil {
		return fmt.Errorf("failed to execute preview command: %s", err)
	}

	// Clean up temporary file if created
	if file != "" {
		return os.Remove(file)
	}

	return nil
}

// getPreviewPaneDimensions attempts to get the actual preview pane size
func getPreviewPaneDimensions() (string, string, error) {
	// Method 1: Try to get terminal size using stty
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	output, err := cmd.Output()
	if err == nil {
		parts := strings.Fields(string(output))
		if len(parts) == 2 {
			// fzf preview typically uses ~40% of terminal width
			// Adjust these percentages based on your fzf configuration
			totalHeight, _ := strconv.Atoi(parts[0])
			totalWidth, _ := strconv.Atoi(parts[1])

			// Assuming default fzf preview layout
			previewHeight := totalHeight - 3               // Account for fzf interface
			previewWidth := int(float64(totalWidth) * 0.5) // Half screen by default

			return strconv.Itoa(previewWidth), strconv.Itoa(previewHeight), nil
		}
	}

	// Method 2: Use tput
	widthCmd := exec.Command("tput", "cols")
	heightCmd := exec.Command("tput", "lines")

	widthOut, errW := widthCmd.Output()
	heightOut, errH := heightCmd.Output()

	if errW == nil && errH == nil {
		width, _ := strconv.Atoi(strings.TrimSpace(string(widthOut)))
		height, _ := strconv.Atoi(strings.TrimSpace(string(heightOut)))

		// Adjust for fzf preview pane (customize based on your layout)
		previewWidth := int(float64(width) * 0.5)
		previewHeight := height - 3

		return strconv.Itoa(previewWidth), strconv.Itoa(previewHeight), nil
	}

	return "", "", fmt.Errorf("could not determine terminal dimensions")
}

// cellsToPixels converts terminal cell dimensions to pixel dimensions
func cellsToPixels(cellWidth, cellHeight string) (string, string) {
	// Default cell dimensions (can be made configurable)
	cellPixelWidth := 9   // Typical terminal cell width in pixels
	cellPixelHeight := 18 // Typical terminal cell height in pixels

	// Check for common terminal environment variables
	if os.Getenv("TERM_PROGRAM") == "iTerm.app" {
		cellPixelWidth = 8
		cellPixelHeight = 16
	}

	width, _ := strconv.Atoi(cellWidth)
	height, _ := strconv.Atoi(cellHeight)

	pixelWidth := width * cellPixelWidth
	pixelHeight := height * cellPixelHeight

	// Apply some padding to ensure image fits well
	pixelWidth = int(float64(pixelWidth) * 0.95)
	pixelHeight = int(float64(pixelHeight) * 0.95)

	return strconv.Itoa(pixelWidth), strconv.Itoa(pixelHeight)
}

// checkSixelSupport detects if the terminal supports sixel graphics
func checkSixelSupport() bool {
	// Check TERM environment variable for known sixel-capable terminals
	term := os.Getenv("TERM")
	sixelTerms := []string{"xterm-256color", "mlterm", "yaft-256color", "foot", "contour", "tmux-256color"}

	for _, st := range sixelTerms {
		if strings.Contains(term, st) {
			return true
		}
	}

	// Check for explicit sixel support environment variable
	if os.Getenv("SIXEL_SUPPORT") == "1" {
		return true
	}
	return false
}

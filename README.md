<div align="center">

# 🖼️ wallhaven-cli

<sub>Search and download wallpapers from [wallhaven](https://wallhaven.cc).</sub>

</div>

## 📦 Installation

### Prerequisites

Before installing `wallhaven-cli`, make sure your system meets the following requirements:

- **Supported platform**:  
  - Linux
  - FreeBSD
- **[Go](https://go.dev/)**:  
  Required to build this project from the source.
- **[fzf](https://github.com/junegunn/fzf?tab=readme-ov-file#installation)**:  
  Used for the selection menu. This is **required**.
- **[chafa](https://hpjansson.org/chafa/)**  
  Used for displaying images in your terminal. This is **required** if you're not using the `kitty` terminal.
- **[kitty](https://github.com/kovidgoyal/kitty)**  
  `kitty` is **optional**.
### Installing

Once all prerequisites are met, you can install `wallhaven-cli` using one of the following methods:

#### 1. Install via `go install`

```bash
go install github.com/ananasmoe/wallhaven-cli/v2/wallhaven@latest
```

#### 2. AUR

You can install `wallhaven-cli` via the AUR:
```bash
yay -S wallhaven-cli
```

If you use a different AUR helper then replace `yay`.

#### 3. Download the Latest Release

Alternatively, you can download the [latest release](https://github.com/ananasmoe/wallhaven-cli/releases/latest) and move the binary to your `~/.local/bin` directory.

---

## 🚀 Usage

[!showcase](https://github.com/user-attachments/assets/42a0d02e-8897-40b1-b79c-ab9c9bb000f4)

### View Available Options

To view all available commands and options, run:

```bash
wallhaven
```

### Edit Configuration

You can customize the settings, such as specifying the folder where downloaded wallpapers will be saved. To edit the configuration, run:

```bash
wallhaven edit
```

This will open the configuration file, where you can set your preferred download directory.

### Search Wallpapers

To search for wallpapers on Wallhaven, use the search command followed by your query:

```bash
wallhaven search [query]
```

For example:

```bash
wallhaven search mashiro shiina
```

### Download Wallpapers

To download a specific wallpaper by its Wallhaven ID, use the following command:

```bash
wallhaven download [id]
```

For example, to download a wallpaper with the ID `l35l5l`:

```bash
wallhaven download l35l5l
```

> [!Tip]
> You can download multiple images at once:
> 
> ```bash
> wallhaven download l35l5l zy759g
> ```
### Download Collections

To download wallpapers from a user’s collection, use the following command:

```bash
wallhaven collection [username]
```

> [!Tip]
> To download all wallpapers from the collection, add the `-a` flag:
> 
> ```bash
> wallhaven collection [username] -a
> ```

---

> [!NOTE]
> I use this project to learn golang. So don't expect the best golang code.

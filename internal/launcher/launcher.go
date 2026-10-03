package launcher

import (
	"bufio"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"github.com/yourname/zompy-launcher/internal/config"
	"github.com/yourname/zompy-launcher/internal/storage"
)

type Minecraft struct {
	cfg   *config.Config
	build storage.Build
	cmd   *exec.Cmd
	mu    sync.Mutex
	done  chan struct{}
}

func New(cfg *config.Config, b storage.Build) *Minecraft {
	return &Minecraft{cfg: cfg, build: b}
}

func (m *Minecraft) Start(logf func(string, any)) error {
	java := m.cfg.JavaPath
	if java == "" {
		java = "javaw"
		if runtime.GOOS != "windows" {
			java = "java"
		}
	}

	args := []string{
		fmt.Sprintf("-Xmx%dG", m.build.Ram),
		fmt.Sprintf("-Xms%dG", max(1, m.build.Ram/2)),
		"-Dfile.encoding=UTF-8",
		"-cp", "libraries/*",
		"net.minecraft.client.main.Main",
		"--version", m.build.Version,
		"--gameDir", m.cfg.GameDir,
		"--assetsDir", m.cfg.GameDir + "/assets",
		"--username", m.cfg.Username,
	}

	cmd := exec.Command(java, args...)
	cmd.Dir = m.cfg.GameDir

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("java start: %w", err)
	}

	m.mu.Lock()
	m.cmd = cmd
	m.done = make(chan struct{})
	m.mu.Unlock()

	go m.pipe(stdout, logf, "stdout")
	go m.pipe(stderr, logf, "stderr")

	go func() {
		_ = cmd.Wait()
		close(m.done)
		logf("game.exit", "Minecraft закрыт")
	}()

	logf("game.start", fmt.Sprintf("%s → %s", m.build.Name, m.build.Version))
	return nil
}

func (m *Minecraft) pipe(r interface{ Read([]byte) (int, error) }, logf func(string, any), tag string) {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		logf("game.log", "["+tag+"] "+line)
	}
}

func (m *Minecraft) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.cmd == nil || m.cmd.Process == nil {
		return nil
	}
	return m.cmd.Process.Kill()
}

func RevealInFileManager(path string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("explorer", strings.ReplaceAll(path, "/", "\\")).Start()
	case "darwin":
		return exec.Command("open", path).Start()
	default:
		return exec.Command("xdg-open", path).Start()
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

# Stocat Commander

A powerful Terminal User Interface (TUI) for easily managing and running actions across the `stocat` microservices architecture. It allows developers to browse repositories, clone them instantly, execute docker containers, run Makefiles, and stream live logs directly inside the terminal.

### Key Concepts

- **Centralized Infrastructure Management**: Set up DBs, Redis, and Consul without leaving your CLI or digging through multiple project folders.
- **Auto-Cloning & Code Refreshing**: Safely clones new projects exactly where they belong (`stocat-asset`, `stocat-gateway`, `stocat-auth`) or automatically runs `git pull` if they already exist, all via simple GUI choices.
- **Integrated Live Logging**: Follow background actions and backend servers in real-time. Features scrollable log history, highlighting, and instant keyword searching.
- **Asynchronous Execution**: Scripts and container startups run in the background with a visual loading spinner, keeping your command hub fully responsive.

---

### Prerequisites

- [Docker & Docker Compose](https://docs.docker.com/get-docker/) (For running backend infrastructure)
- [Homebrew](https://brew.sh/) (On MacOS/Linux to install Go automatically if it's missing)

---

### Basic Usage

1. Open your terminal and navigate to the `stocat-commander` directory.
2. Run the application via the Makefile:

   ```bash
   make run
   ```

   _The `Makefile` will automatically verify your Go installation and try to install it via Homebrew (`brew`) if you're missing it._

3. Use the **Up/Down Arrow Keys** to navigate menus.
4. Press **Enter** to jump into projects or execute commands.
5. In the log viewer, hit **`/`** to search through live and historical logs, and **`Esc`** to back out without closing servers.
6. **`q`** or **`Ctrl+C`** securely quits the app at any time.

Enjoy faster microservice administration!

package commands

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"

	"github.com/onuse/tasks/internal/store"
)

func Serve(args []string) error {
	// Parse flags
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	port := fs.Int("port", 8080, "Port to serve on")
	noBrowser := fs.Bool("no-browser", false, "Don't open browser automatically")
	fs.Parse(args)

	// Find task root
	rootDir, err := store.FindTaskRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Run 'task init' to initialize task tracking in this repository.\n")
		os.Exit(3)
	}

	s := store.New(rootDir)

	// Setup HTTP handlers
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveIndex(w, r, s)
	})

	http.HandleFunc("/api/tasks", func(w http.ResponseWriter, r *http.Request) {
		serveTasksAPI(w, r, s)
	})

	http.HandleFunc("/api/task/", func(w http.ResponseWriter, r *http.Request) {
		serveTaskAPI(w, r, s)
	})

	addr := fmt.Sprintf(":%d", *port)
	url := fmt.Sprintf("http://localhost:%d", *port)

	fmt.Printf("Starting task server on %s\n", url)
	fmt.Println("Press Ctrl+C to stop")

	// Open browser
	if !*noBrowser {
		go func() {
			time.Sleep(500 * time.Millisecond)
			openBrowser(url)
		}()
	}

	return http.ListenAndServe(addr, nil)
}

func serveIndex(w http.ResponseWriter, r *http.Request, s *store.Store) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(htmlTemplate))
}

func serveTasksAPI(w http.ResponseWriter, r *http.Request, s *store.Store) {
	index, err := s.ReadIndex()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(index.Tasks)
}

func serveTaskAPI(w http.ResponseWriter, r *http.Request, s *store.Store) {
	// Extract task ID from URL path
	idStr := r.URL.Path[len("/api/task/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	task, err := s.ReadTask(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Printf("Failed to open browser: %v", err)
	}
}

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Task Manager</title>
    <style>
        * {
            box-sizing: border-box;
            margin: 0;
            padding: 0;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: #f5f5f5;
            padding: 20px;
        }

        .container {
            max-width: 1400px;
            margin: 0 auto;
        }

        h1 {
            margin-bottom: 30px;
            color: #333;
        }

        .board {
            display: flex;
            gap: 20px;
            overflow-x: auto;
            padding-bottom: 20px;
        }

        .column {
            flex: 1;
            min-width: 300px;
            background: #fff;
            border-radius: 8px;
            padding: 15px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }

        .column-header {
            font-weight: 600;
            font-size: 14px;
            text-transform: uppercase;
            color: #666;
            margin-bottom: 15px;
            padding-bottom: 10px;
            border-bottom: 2px solid #e0e0e0;
        }

        .task-card {
            background: #fafafa;
            border: 1px solid #e0e0e0;
            border-radius: 6px;
            padding: 12px;
            margin-bottom: 10px;
            cursor: pointer;
            transition: all 0.2s;
        }

        .task-card:hover {
            box-shadow: 0 4px 8px rgba(0,0,0,0.15);
            transform: translateY(-2px);
        }

        .task-id {
            font-size: 11px;
            color: #999;
            font-weight: 600;
        }

        .task-title {
            font-size: 14px;
            color: #333;
            margin-top: 5px;
            font-weight: 500;
        }

        .task-date {
            font-size: 11px;
            color: #999;
            margin-top: 8px;
        }

        .status-backlog { border-left: 4px solid #9e9e9e; }
        .status-next { border-left: 4px solid #FFC107; }
        .status-active { border-left: 4px solid #2196F3; }
        .status-blocked { border-left: 4px solid #ff9800; }
        .status-done { border-left: 4px solid #4caf50; }
        .status-cancelled { border-left: 4px solid #f44336; }

        .modal {
            display: none;
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: rgba(0,0,0,0.5);
            z-index: 1000;
        }

        .modal-content {
            position: relative;
            background: white;
            max-width: 800px;
            margin: 50px auto;
            padding: 30px;
            border-radius: 8px;
            max-height: 80vh;
            overflow-y: auto;
        }

        .modal-close {
            position: absolute;
            top: 15px;
            right: 15px;
            font-size: 24px;
            cursor: pointer;
            color: #999;
        }

        .modal-close:hover {
            color: #333;
        }

        .task-detail h2 {
            margin-bottom: 20px;
            color: #333;
        }

        .task-meta {
            display: grid;
            grid-template-columns: repeat(2, 1fr);
            gap: 15px;
            margin-bottom: 20px;
            padding: 15px;
            background: #f9f9f9;
            border-radius: 6px;
        }

        .meta-item {
            font-size: 13px;
        }

        .meta-label {
            color: #666;
            font-weight: 600;
            margin-bottom: 4px;
        }

        .meta-value {
            color: #333;
        }

        .section {
            margin-top: 25px;
        }

        .section-title {
            font-size: 14px;
            font-weight: 600;
            color: #666;
            margin-bottom: 10px;
            text-transform: uppercase;
        }

        .description {
            line-height: 1.6;
            color: #333;
        }

        .links-list, .notes-list {
            list-style: none;
        }

        .link-item, .note-item {
            padding: 10px;
            background: #f9f9f9;
            border-radius: 4px;
            margin-bottom: 8px;
            font-size: 13px;
        }

        .link-type {
            font-weight: 600;
            color: #2196F3;
        }

        .note-meta {
            color: #666;
            font-size: 12px;
            margin-bottom: 5px;
        }

        .note-text {
            color: #333;
        }

        .refresh-btn {
            position: fixed;
            bottom: 30px;
            right: 30px;
            background: #2196F3;
            color: white;
            border: none;
            padding: 15px 25px;
            border-radius: 50px;
            cursor: pointer;
            font-size: 14px;
            font-weight: 600;
            box-shadow: 0 4px 12px rgba(33,150,243,0.4);
            transition: all 0.2s;
        }

        .refresh-btn:hover {
            background: #1976D2;
            box-shadow: 0 6px 16px rgba(33,150,243,0.6);
        }

        .empty-column {
            color: #999;
            font-size: 13px;
            text-align: center;
            padding: 20px;
        }

        .controls {
            margin-bottom: 20px;
            display: flex;
            gap: 15px;
            align-items: center;
        }

        .search-box {
            flex: 1;
            padding: 10px 15px;
            border: 1px solid #ddd;
            border-radius: 6px;
            font-size: 14px;
        }

        .search-box:focus {
            outline: none;
            border-color: #2196F3;
        }

        .view-toggle {
            display: flex;
            gap: 10px;
        }

        .view-btn {
            padding: 10px 20px;
            border: 1px solid #ddd;
            background: white;
            border-radius: 6px;
            cursor: pointer;
            font-size: 14px;
            transition: all 0.2s;
        }

        .view-btn.active {
            background: #2196F3;
            color: white;
            border-color: #2196F3;
        }

        .view-btn:hover {
            border-color: #2196F3;
        }

        .list-view {
            background: white;
            border-radius: 8px;
            padding: 20px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }

        .list-task {
            padding: 15px;
            border-bottom: 1px solid #e0e0e0;
            cursor: pointer;
            transition: background 0.2s;
        }

        .list-task:hover {
            background: #f9f9f9;
        }

        .list-task:last-child {
            border-bottom: none;
        }

        .list-task-header {
            display: flex;
            align-items: center;
            gap: 15px;
            margin-bottom: 5px;
        }

        .list-task-status {
            padding: 4px 8px;
            border-radius: 4px;
            font-size: 11px;
            font-weight: 600;
            text-transform: uppercase;
        }

        .list-task-status.backlog { background: #9e9e9e; color: white; }
        .list-task-status.next { background: #FFC107; color: white; }
        .list-task-status.active { background: #2196F3; color: white; }
        .list-task-status.blocked { background: #ff9800; color: white; }
        .list-task-status.done { background: #4caf50; color: white; }
        .list-task-status.cancelled { background: #f44336; color: white; }

        .list-task-id {
            font-size: 12px;
            color: #999;
            font-weight: 600;
        }

        .list-task-title {
            font-size: 16px;
            font-weight: 500;
            color: #333;
            flex: 1;
        }

        .list-task-date {
            font-size: 12px;
            color: #999;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>📋 Task Manager</h1>

        <div class="controls">
            <input type="text" id="searchBox" class="search-box" placeholder="Search tasks..." onkeyup="filterTasks()">
            <div class="view-toggle">
                <button class="view-btn active" id="boardViewBtn" onclick="setView('board')">Board</button>
                <button class="view-btn" id="listViewBtn" onclick="setView('list')">List</button>
            </div>
        </div>

        <div class="board" id="board"></div>
        <div class="list-view" id="listView" style="display: none;"></div>
    </div>

    <button class="refresh-btn" onclick="loadTasks()">🔄 Refresh</button>

    <div class="modal" id="taskModal">
        <div class="modal-content">
            <span class="modal-close" onclick="closeModal()">&times;</span>
            <div id="taskDetail"></div>
        </div>
    </div>

    <script>
        let tasks = [];
        let filteredTasks = [];
        let currentView = 'board';

        async function loadTasks() {
            try {
                const response = await fetch('/api/tasks');
                tasks = await response.json();
                filterTasks();
            } catch (error) {
                console.error('Failed to load tasks:', error);
            }
        }

        function filterTasks() {
            const searchTerm = document.getElementById('searchBox').value.toLowerCase();

            if (searchTerm === '') {
                filteredTasks = tasks;
            } else {
                filteredTasks = tasks.filter(task =>
                    task.title.toLowerCase().includes(searchTerm) ||
                    task.id.toString().includes(searchTerm)
                );
            }

            if (currentView === 'board') {
                renderBoard();
            } else {
                renderList();
            }
        }

        function setView(view) {
            currentView = view;

            if (view === 'board') {
                document.getElementById('board').style.display = 'flex';
                document.getElementById('listView').style.display = 'none';
                document.getElementById('boardViewBtn').classList.add('active');
                document.getElementById('listViewBtn').classList.remove('active');
                renderBoard();
            } else {
                document.getElementById('board').style.display = 'none';
                document.getElementById('listView').style.display = 'block';
                document.getElementById('boardViewBtn').classList.remove('active');
                document.getElementById('listViewBtn').classList.add('active');
                renderList();
            }
        }

        function renderBoard() {
            const statuses = ['backlog', 'next', 'active', 'blocked', 'done', 'cancelled'];
            const statusNames = {
                'backlog': 'Backlog',
                'next': 'Next',
                'active': 'Active',
                'blocked': 'Blocked',
                'done': 'Done',
                'cancelled': 'Cancelled'
            };

            const board = document.getElementById('board');
            board.innerHTML = '';

            statuses.forEach(status => {
                const column = document.createElement('div');
                column.className = 'column';

                const header = document.createElement('div');
                header.className = 'column-header';
                const statusTasks = filteredTasks.filter(t => t.status === status);
                header.textContent = statusNames[status] + ' (' + statusTasks.length + ')';
                column.appendChild(header);

                if (statusTasks.length === 0) {
                    const empty = document.createElement('div');
                    empty.className = 'empty-column';
                    empty.textContent = 'No tasks';
                    column.appendChild(empty);
                } else {
                    statusTasks.forEach(task => {
                        const card = createTaskCard(task);
                        column.appendChild(card);
                    });
                }

                board.appendChild(column);
            });
        }

        function createTaskCard(task) {
            const card = document.createElement('div');
            card.className = 'task-card status-' + task.status;
            card.onclick = () => showTask(task.id);

            const id = document.createElement('div');
            id.className = 'task-id';
            id.textContent = '#' + task.id;

            const title = document.createElement('div');
            title.className = 'task-title';
            title.textContent = task.title;

            const date = document.createElement('div');
            date.className = 'task-date';
            date.textContent = formatDate(task.updated);

            card.appendChild(id);
            card.appendChild(title);
            card.appendChild(date);

            return card;
        }

        async function showTask(id) {
            try {
                const response = await fetch('/api/task/' + id);
                const task = await response.json();
                renderTaskDetail(task);
                document.getElementById('taskModal').style.display = 'block';
            } catch (error) {
                console.error('Failed to load task:', error);
            }
        }

        function renderTaskDetail(task) {
            const detail = document.getElementById('taskDetail');

            let html = '<div class="task-detail">';
            html += '<h2>#' + task.id + ': ' + task.title + '</h2>';

            html += '<div class="task-meta">';
            html += '<div class="meta-item"><div class="meta-label">Status</div><div class="meta-value">' + task.status + '</div></div>';
            html += '<div class="meta-item"><div class="meta-label">Created</div><div class="meta-value">' + formatDate(task.created) + '</div></div>';
            html += '<div class="meta-item"><div class="meta-label">Updated</div><div class="meta-value">' + formatDate(task.updated) + '</div></div>';
            if (task.tags && task.tags.length > 0) {
                html += '<div class="meta-item"><div class="meta-label">Tags</div><div class="meta-value">' + task.tags.join(', ') + '</div></div>';
            }
            html += '</div>';

            if (task.description) {
                html += '<div class="section">';
                html += '<div class="section-title">Description</div>';
                html += '<div class="description">' + escapeHtml(task.description) + '</div>';
                html += '</div>';
            }

            if (task.links && task.links.length > 0) {
                html += '<div class="section">';
                html += '<div class="section-title">Links</div>';
                html += '<ul class="links-list">';
                task.links.forEach(link => {
                    html += '<li class="link-item">';
                    html += '<span class="link-type">' + link.type + '</span> #' + link.target_id;
                    if (link.label) {
                        html += ' <em>(' + escapeHtml(link.label) + ')</em>';
                    }
                    html += '</li>';
                });
                html += '</ul></div>';
            }

            if (task.notes && task.notes.length > 0) {
                html += '<div class="section">';
                html += '<div class="section-title">Notes</div>';
                html += '<ul class="notes-list">';
                task.notes.forEach(note => {
                    html += '<li class="note-item">';
                    html += '<div class="note-meta">' + formatDate(note.timestamp) + ' - ' + note.author + '</div>';
                    html += '<div class="note-text">' + escapeHtml(note.text) + '</div>';
                    html += '</li>';
                });
                html += '</ul></div>';
            }

            html += '</div>';
            detail.innerHTML = html;
        }

        function closeModal() {
            document.getElementById('taskModal').style.display = 'none';
        }

        function formatDate(dateStr) {
            const date = new Date(dateStr);
            return date.toLocaleString();
        }

        function escapeHtml(text) {
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        }

        function renderList() {
            const listView = document.getElementById('listView');
            listView.innerHTML = '';

            if (filteredTasks.length === 0) {
                listView.innerHTML = '<div class="empty-column">No tasks found</div>';
                return;
            }

            // Sort by ID descending (newest first)
            const sortedTasks = [...filteredTasks].sort((a, b) => b.id - a.id);

            sortedTasks.forEach(task => {
                const taskDiv = document.createElement('div');
                taskDiv.className = 'list-task';
                taskDiv.onclick = () => showTask(task.id);

                const header = document.createElement('div');
                header.className = 'list-task-header';

                const statusSpan = document.createElement('span');
                statusSpan.className = 'list-task-status ' + task.status;
                statusSpan.textContent = task.status;

                const idSpan = document.createElement('span');
                idSpan.className = 'list-task-id';
                idSpan.textContent = '#' + task.id;

                const titleSpan = document.createElement('span');
                titleSpan.className = 'list-task-title';
                titleSpan.textContent = task.title;

                const dateSpan = document.createElement('span');
                dateSpan.className = 'list-task-date';
                dateSpan.textContent = formatDate(task.updated);

                header.appendChild(statusSpan);
                header.appendChild(idSpan);
                header.appendChild(titleSpan);
                header.appendChild(dateSpan);

                taskDiv.appendChild(header);
                listView.appendChild(taskDiv);
            });
        }

        // Close modal when clicking outside
        window.onclick = function(event) {
            const modal = document.getElementById('taskModal');
            if (event.target == modal) {
                closeModal();
            }
        }

        // Load tasks on page load
        loadTasks();

        // Auto-refresh every 5 seconds
        setInterval(loadTasks, 5000);
    </script>
</body>
</html>
`

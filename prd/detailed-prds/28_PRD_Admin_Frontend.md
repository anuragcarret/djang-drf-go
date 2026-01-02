# PRD: Admin Site Frontend (Embedded React App)

## 1. Objective
Provide a modern, premium, and responsive Admin Interface for `django-drf-go` that is served directly by the Go backend without requiring a separate Node.js server for deployment.

## 2. Technical Stack
- **Framework**: React 18+ (SPA)
- **Build Tool**: Vite (Fast, efficient bundling)
- **Styling**: TailwindCSS (Premium UI, dark mode support)
- **Integration**: Go `embed` package to bundle the `dist/` folder into the binary.
- **Routing**: `react-router-dom` (Client-side routing).
- **API Client**: Native `fetch` talking to the backend's `/admin/*` JSON APIs.

## 3. Architecture
The frontend source will live in `django_drf_go/admin/ui`.
The build output will be targeted to `django_drf_go/admin/dist`.

```
django_drf_go/
  admin/
    ui/          # React Source
      src/
      package.json
      vite.config.js
    dist/        # Generated Assets (Checked in or generated)
    static.go    # Embeds dist/
    admin.go     # Serves static files
```

## 4. Key Features

### 4.1 Dashboard (`/admin/`)
- Display all registered Apps and Models.
- Grouped by App Label.
- Stats cards (e.g., "Total Users").

### 4.2 Model List View (`/admin/:app/:model`)
- Fetch data from generic `ListModelView` endpoint.
- **Components**:
  - Dynamic Data Table.
  - Pagination controls.
  - Search bar (future).
  - Filters sidebar (future).

### 4.3 Design Requirements
- **Premium Aesthetics**: Glassmorphism effects, smooth transitions, Inter font.
- **Responsive**: Mobile-friendly layout.
- **Dark Mode**: Toggle support.

## 5. Development Workflow (TDD)
1.  **Backend Serving Test**: Ensure Go can serve a simple HTML file from `dist`.
2.  **React Setup**: Initialize Vite app in `admin/ui`.
3.  **API Integration**: Build the Dashboard component and verify it renders data from the API.

## 6. Deployment
- The user runs `npm run build` in `admin/ui`.
- The `dist` folder is updated.
- The Go binary embeds `dist` and serves it easily.

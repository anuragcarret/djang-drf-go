# PRD: Template-based Admin Panel (Go Templates)

## Overview
Re-implement the `django-drf-go` Admin Site using Go's standard `html/template` package. The goal is to provide a "zero-config" admin panel that is automatically served by the Go binary without needing a separate frontend build process.

## Objectives
1.  **Simplicity**: No Node.js/React dependency.
2.  **Generalizability**: Automatically handle any model registered via `admin.Register[T]()`.
3.  **Composability**: Easy to include in any `django-drf-go` project.
4.  **TDD Driven**: Test routing, template rendering, and form submission.

## Core Features
1.  **Base Layout**: Shared header, navigation (sidebar with apps/models), and content area.
2.  **Dashboard**: Landing page listing all available apps and models.
3.  **List View**: Data table for a specific model with basic sorting/filtering.
4.  **Detail View (Add/Edit)**: Form generation based on model fields for creating or updating records.
5.  **Actions**: Support for bulk actions (e.g., Delete selected).

## UI/UX Goals
- Modern, clean aesthetic (using a lightweight CSS framework or vanilla CSS).
- Responsive design.
- Clear breadcrumbs and navigation.

## Technical Strategy
- **Templates**: Organized in `admin/templates/`.
- **Views**: Go handlers that fetch data via `QuerySet` and render templates.
- **Forms**: Dynamic form generation using reflection to inspect model fields and tags.
- **Static Assets**: CSS/JS embedded using `//go:embed`.

## Roadmap
1.  **Phase 27**: Setup & Base Layout (TDD: Test basic HTML rendering).
2.  **Phase 28**: Dashboard View (Lists apps/models).
3.  **Phase 29**: Generic List View (Data table with reflection).
4.  **Phase 30**: Detail View & Form Handling (Add/Edit logic).
5.  **Phase 31**: Polish & Demo Integration.

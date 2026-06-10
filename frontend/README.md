# TMS Frontend

Frontend application for the Test Management System (TMS).

The project is a single-page application built with:

- React
- TypeScript
- Vite
- Native Fetch API

The application provides a complete workflow for managing:

- Authentication
- Projects
- Folder structures
- Test cases
- Test case versions
- Test suites
- Test execution runs
- Execution results
- Bug references

The project does not use any router or UI component library. All UI elements are built with React components and custom CSS.

## Getting Started

### Requirements

Make sure you have installed:

- Node.js
- npm

## Installation

Install dependencies:

```bash
npm install
```

## Development

Run the development server:

```bash
npm run dev
```

The development server proxies API requests:

```text
/api -> localhost:8080
```

## Build

Create a production build:

```bash
npm run build
```

Preview the production build:

```bash
npm run preview
```

## Project Structure

```text
src/
├── main.tsx
│   Application entry point
│
├── App.tsx
│   Main application component
│   Handles authentication state,
│   project selection and tab navigation
│
├── api.ts
│   Central API client
│   Contains all backend requests
│   and TypeScript models
│
├── index.css
│   Global styles and base UI styles
│
├── App.css
│   Application layout styles
│
└── components/
    ├── Login.tsx
    │   User registration and login
    │
    ├── Projects.tsx
    │   Project list and creation
    │
    ├── Folders.tsx
    │   Folder tree management
    │
    ├── TestCases.tsx
    │   Test case creation,
    │   editing and version history
    │
    ├── Suites.tsx
    │   Test suite management
    │
    └── Runs.tsx
        Test execution management
```

## Application Flow

The application follows this flow:

```text
Login
  ↓
Select Project
  ↓
Manage Test Data
  ↓
Execute Test Runs
```

A user must be authenticated and select a project before accessing the main features.

## Authentication

Authentication is handled using JWT tokens.

After successful login:

- The token is stored in `localStorage`
- The token is attached automatically to API requests
- The user can refresh the page without logging in again

Logout:

- Removes the stored token
- Clears application state
- Returns the user to the login page

## API Layer

All backend communication is located in:

```text
src/api.ts
```

Components do not call `fetch` directly.

The API layer handles:

- Request creation
- JSON serialization
- Authorization headers
- Error handling
- Response parsing

Example:

```ts
api.listProjects()
api.createCase()
api.markResult()
```

This keeps API logic separated from UI components.

## Main Features

### Projects

Projects are the top-level container.

Users can:

- View projects
- Create new projects
- Select an active project

The selected project controls all following operations.

### Folder Management

Folders organize test cases.

Features:

- Nested folder tree
- Create root folders
- Create child folders
- Select parent folder

Example:

```text
Project

├── Authentication
│   ├── Login Tests
│   └── Registration Tests
│
└── Payments
```

Folders are displayed recursively.

### Test Cases

Test cases are the main testing objects.

Supported operations:

- Create test cases
- View test cases
- Edit test cases
- View previous versions

A test case contains:

- Title
- Description
- Preconditions
- Steps
- Expected result
- Priority
- Status

### Test Case Versioning

Test cases use version history.

Editing a test case creates a new version instead of replacing the old one.

Example:

```text
Version 1
    |
    ↓
Version 2
    |
    ↓
Version 3
```

Previous versions remain available and can be viewed anytime.

### Test Suites

Suites are groups of test cases.

Users can:

- Create suites
- Add test cases
- Remove test cases
- View suite content

Example:

```text
Regression Suite

├── Login Test
├── Checkout Test
└── Profile Test
```

### Test Runs

A run represents execution of a test suite.

Users can:

- Create runs
- Execute test cases
- Update results
- Track execution progress
- Attach bugs

Supported result statuses:

```text
passed
failed
skipped
blocked
not_run
```

### Run Snapshots

When a run is created, the current test case versions are saved.

Example:

```text
Test Case Version
        |
        ↓
Run Snapshot
```

Future test case edits do not affect existing runs.

This keeps execution history accurate.

### Bug Tracking

Execution results can contain linked bugs.

Bug information includes:

- Tracker name
- Bug ID
- Optional URL

Example:

```text
Tracker: Jira
ID: BUG-123
URL: https://example.com/BUG-123
```

## Styling

The project uses custom CSS.

### index.css

Contains:

- Global reset
- Base typography
- Form styles
- Button styles
- Dark mode support

### App.css

Contains:

- Layout
- Panels
- Tabs
- Lists
- Tables
- Status badges
- Responsive behavior

## Component Responsibilities

| Component | Purpose |
|---|---|
| Login | Authentication |
| Projects | Project selection |
| Folders | Folder hierarchy |
| TestCases | Test case lifecycle |
| Suites | Suite management |
| Runs | Test execution |

## State Management

The application uses React built-in state.

Main application state:

```text
App.tsx
```

Handles:

- Authentication status
- Current project
- Active tab

Each feature component manages its own local state.

## Development Guidelines

- Add new API calls only inside `api.ts`
- Keep backend models synchronized with TypeScript interfaces
- Do not call `fetch` directly inside components
- Keep components focused on UI behavior
- Preserve test case version history when editing cases

## Technology Stack

| Technology | Purpose |
|---|---|
| React | UI framework |
| TypeScript | Type safety |
| Vite | Development and build tool |
| CSS | Styling |
| Fetch API | Backend communication |

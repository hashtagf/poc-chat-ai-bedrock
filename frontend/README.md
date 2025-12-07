# Chat UI - Bedrock Agent Core

Vue 3 chat interface for Amazon Bedrock Agent Core with knowledge base integration.

## Project Structure

```
frontend/
├── src/
│   ├── components/      # Vue components
│   ├── composables/     # Business logic (Vue composables)
│   ├── types/           # TypeScript type definitions
│   ├── tests/           # Unit and property-based tests
│   ├── App.vue          # Root component
│   ├── main.ts          # Application entry point
│   └── style.css        # Global styles with Tailwind
├── index.html           # HTML entry point
├── vite.config.ts       # Vite configuration
├── vitest.config.ts     # Vitest test configuration
├── tailwind.config.js   # Tailwind CSS configuration
└── package.json         # Dependencies and scripts
```

## Setup

Install dependencies:

```bash
npm install
```

## Development

Start the development server:

```bash
npm run dev
```

The application will be available at `http://localhost:3000`

## Testing

Run tests once:

```bash
npm test
```

Run tests in watch mode:

```bash
npm run test:watch
```

Run tests with coverage:

```bash
npm run test:coverage
```

## Build

Build for production:

```bash
npm run build
```

Preview production build:

```bash
npm run preview
```

## Technology Stack

- **Vue 3** with Composition API
- **TypeScript** for type safety
- **Vite** for fast development and building
- **Tailwind CSS** for styling
- **Vitest** for unit testing
- **fast-check** for property-based testing
- **Vue Test Utils** for component testing

## Architecture

This project follows hexagonal architecture principles:

- **Components**: UI layer (Vue components)
- **Composables**: Business logic layer (framework-agnostic)
- **Types**: Domain models and interfaces

The UI communicates with a Go backend that handles Bedrock Agent Core integration.

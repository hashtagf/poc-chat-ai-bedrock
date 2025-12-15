# Chat UI - Bedrock Agent Core ✅ FULLY FUNCTIONAL

Vue 3 chat interface for Amazon Bedrock Agent Core with S3 Vectors knowledge base integration.

**Status**: Production ready with real-time streaming, citations, and comprehensive testing.

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

The application will be available at `http://localhost:5173`

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

## Features ✅

- ✅ **Real-time Streaming**: Incremental display of AI responses
- ✅ **Knowledge Base Citations**: S3 Vectors citations with confidence scores
- ✅ **Session Management**: Multiple conversation sessions
- ✅ **Error Handling**: User-friendly error messages with retry
- ✅ **WebSocket Communication**: Real-time bidirectional messaging
- ✅ **Responsive Design**: Tailwind CSS with mobile support
- ✅ **TypeScript**: Full type safety throughout
- ✅ **Testing**: Unit tests and property-based testing

## Architecture

This project follows hexagonal architecture principles with complete integration:

- **Components**: UI layer (Vue 3 Composition API)
- **Composables**: Business logic layer (WebSocket, session management)
- **Types**: Domain models and interfaces (TypeScript)
- **Integration**: Real-time communication with Go backend and Bedrock Agent Core

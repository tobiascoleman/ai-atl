# AI-ATL NFL Platform - Frontend

Modern Next.js 14 frontend for the AI-powered NFL fantasy platform.

## 🚀 Quick Start

### Prerequisites
- Node.js 18+ 
- npm or yarn

### Installation

```bash
# Install dependencies
npm install

# Copy environment variables
cp .env.local.example .env.local

# Edit .env.local and set API URL
# NEXT_PUBLIC_API_URL=http://localhost:8080

# Run development server
npm run dev
```

The app will be available at `http://localhost:3000`

## 📁 Project Structure

```
frontend/
├── app/                    # Next.js 14 app directory
│   ├── (auth)/            # Auth pages (login, register)
│   ├── dashboard/         # Dashboard pages
│   │   ├── chat/         # AI Chatbot
│   │   ├── insights/     # Game Script Predictor
│   │   ├── players/      # Player browser
│   │   └── trades/       # Trade analyzer
│   ├── layout.tsx        # Root layout
│   └── page.tsx          # Landing page
├── lib/
│   ├── api/              # API client & endpoints
│   ├── stores/           # Zustand state management
│   └── utils.ts          # Utility functions
├── types/
│   └── api.ts            # TypeScript types
└── components/           # Reusable components (future)
```

## 🎯 Key Features

### 1. AI Game Script Predictor (`/dashboard/insights`)
- Predicts quarter-by-quarter game flow
- Shows player impact predictions
- Confidence scoring
- **This is your main differentiator!**

### 2. AI Chatbot (`/dashboard/chat`)
- Context-aware fantasy advice
- Natural language interface
- Personalized recommendations
- Quick question templates

### 3. Trade Analyzer (`/dashboard/trades`)
- AI-powered trade evaluation
- Fairness scoring
- Grade for each team
- Detailed analysis

### 4. Player Browser (`/dashboard/players`)
- EPA-based metrics
- Filter by position/team
- Search functionality
- Advanced stats display

## 🔌 API Integration

The frontend connects to the Go API backend. All API calls are in `lib/api/`:

```typescript
// Example: Chat with AI
import { chatbotAPI } from '@/lib/api/chatbot'

const response = await chatbotAPI.ask("Who should I start?")
```

## 🎨 Styling

- **Framework**: Tailwind CSS
- **Icons**: Lucide React
- **Animations**: Framer Motion (included)

## 🔐 Authentication

Authentication uses JWT tokens stored in localStorage:

```typescript
import { useAuthStore } from '@/lib/stores/authStore'

const { user, isAuthenticated, logout } = useAuthStore()
```

## 📦 Dependencies

### Core
- `next@14.0.4` - React framework
- `react@18.2.0` - UI library
- `typescript@5.3.3` - Type safety

### API & State
- `axios@1.6.2` - HTTP client
- `swr@2.2.4` - Data fetching (for future use)
- `zustand@4.4.7` - State management

### UI & Visualization
- `lucide-react@0.303.0` - Icons
- `recharts@2.10.3` - Charts (for future use)
- `framer-motion@10.16.16` - Animations (for future use)

## 🚀 Building for Production

```bash
# Build the application
npm run build

# Start production server
npm start
```

## 🎯 Environment Variables

Create `.env.local`:

```bash
NEXT_PUBLIC_API_URL=http://localhost:8080
```

For production, update to your deployed API URL.

## 🧪 Testing the App

1. Start the backend API (`go run cmd/api/main.go`)
2. Start the frontend (`npm run dev`)
3. Register a new account at `/register`
4. Test features:
   - Chat: `/dashboard/chat`
   - Game Script: `/dashboard/insights`
   - Trades: `/dashboard/trades`
   - Players: `/dashboard/players`

## 📝 Development Tips

### Adding New Pages

Create a new file in `app/dashboard/`:

```typescript
// app/dashboard/new-page/page.tsx
'use client'

export default function NewPage() {
  return <div>New Page Content</div>
}
```

### Adding API Endpoints

Add to `lib/api/`:

```typescript
// lib/api/new-feature.ts
import apiClient from './client'

export const newFeatureAPI = {
  getData: async () => {
    const { data } = await apiClient.get('/endpoint')
    return data
  }
}
```

### Using the API Client

The API client automatically:
- Adds JWT tokens to requests
- Handles 401 errors (redirects to login)
- Works with the Go backend

## 🎨 Customization

### Colors

Edit `tailwind.config.ts`:

```typescript
colors: {
  primary: {
    DEFAULT: '#1e40af', // Your brand color
  }
}
```

### Layout

Edit `app/dashboard/layout.tsx` to modify the navigation.

## 🐛 Common Issues

### API Connection Errors

Check that:
1. Backend is running on port 8080
2. `NEXT_PUBLIC_API_URL` is set correctly
3. CORS is enabled in backend

### Authentication Issues

Clear localStorage and login again:

```javascript
localStorage.clear()
```

### Build Errors

Clear Next.js cache:

```bash
rm -rf .next
npm run dev
```

## 📱 Mobile Support

The app is responsive but optimized for desktop. For hackathon demos:
- Present on laptop/desktop
- Use browser dev tools to test mobile view
- Focus on functionality over mobile UX

## 🔗 Useful Links

- [Next.js Docs](https://nextjs.org/docs)
- [Tailwind CSS](https://tailwindcss.com/docs)
- [Lucide Icons](https://lucide.dev/icons)

## 🏆 Demo Tips

For your hackathon presentation:

1. **Start with Chat** - Most engaging feature
2. **Show Game Script Predictor** - Your main differentiator
3. **Demonstrate Trade Analyzer** - Show AI in action
4. **Browse Players** - Showcase EPA metrics

---

**Built for ATL Hackathon 2025** 🚀


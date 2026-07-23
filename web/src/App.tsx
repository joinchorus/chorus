import React, { useState, useEffect } from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Navbar } from './components/Navbar';
import { Footer } from './components/Footer';
import { Home } from './pages/Home';
import { BoardDetail } from './pages/BoardDetail';
import { CreateThread } from './pages/CreateThread';
import { ThreadDetail } from './pages/ThreadDetail';
import { NotFound } from './pages/NotFound';
import { OnboardingModal } from './components/OnboardingModal';
import './App.css';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: 1,
    },
  },
});

export const App: React.FC = () => {
  const [isOnboardingOpen, setIsOnboardingOpen] = useState(false);

  useEffect(() => {
    const onboarded = localStorage.getItem('chorus_onboarded');
    if (!onboarded) {
      setIsOnboardingOpen(true);
    }
  }, []);

  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <div className="app-shell">
          <Navbar onOpenOnboarding={() => setIsOnboardingOpen(true)} />
          <main className="app-main">
            <div className="container">
              <Routes>
                <Route path="/" element={<Home />} />
                <Route path="/board/:slug" element={<BoardDetail />} />
                <Route path="/new" element={<CreateThread />} />
                <Route path="/thread/:id" element={<ThreadDetail />} />
                <Route path="*" element={<NotFound />} />
              </Routes>
            </div>
          </main>
          <Footer />
          <OnboardingModal
            isOpen={isOnboardingOpen}
            onClose={() => setIsOnboardingOpen(false)}
          />
        </div>
      </BrowserRouter>
    </QueryClientProvider>
  );
};

export default App;

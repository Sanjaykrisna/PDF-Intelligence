import React, { useState, useEffect } from 'react';
import Conversation from './pages/Conversation';
import OnboardingModal from './components/OnboardingModal';
import { ThemeProvider } from './context/ThemeContext';

export default function App() {
  const [showOnboarding, setShowOnboarding] = useState(false);

  useEffect(() => {
    // Check if user has seen onboarding
    const hasSeenOnboarding = localStorage.getItem('rag_has_seen_onboarding');
    if (!hasSeenOnboarding) {
      setShowOnboarding(true);
    }
  }, []);

  const handleOnboardingComplete = () => {
    localStorage.setItem('rag_has_seen_onboarding', 'true');
    setShowOnboarding(false);
  };

  return (
    <ThemeProvider>
      <div className="w-full min-h-screen bg-bg-surface dark:bg-[#0c0d12] transition-colors duration-200">
        <Conversation
          onOpenOnboarding={() => setShowOnboarding(true)}
        />

        {showOnboarding && (
          <OnboardingModal
            onComplete={handleOnboardingComplete}
            onClose={() => setShowOnboarding(false)}
          />
        )}
      </div>
    </ThemeProvider>
  );
}


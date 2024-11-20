import AuthForm from '@/components/auth-form';
import { Toaster } from '@/components/ui/sonner';

function App() {
  return (
    <div className="items-center justify-center p-4">
      <AuthForm />
      <Toaster position="top-center" />
    </div>
  );
}

export default App;

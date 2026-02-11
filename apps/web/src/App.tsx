import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { Toaster } from 'sonner';
import { Layout } from './components';
import { Login, Register, Generate, Gallery, Credits, Members, Providers } from './pages';
import { useAuthStore } from './stores';

function ProtectedRoute({ children, requireAdmin = false }: { children: React.ReactNode; requireAdmin?: boolean }) {
  const { isAuthenticated, user } = useAuthStore();
  
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }
  
  if (requireAdmin && user?.role !== 'admin') {
    return <Navigate to="/generate" replace />;
  }
  
  return <>{children}</>;
}

function PublicRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useAuthStore();
  
  if (isAuthenticated) {
    return <Navigate to="/generate" replace />;
  }
  
  return <>{children}</>;
}

function App() {
  return (
    <BrowserRouter>
      <Toaster position="top-right" richColors />
      <Routes>
        {/* Public Routes */}
        <Route
          path="/login"
          element={
            <PublicRoute>
              <Login />
            </PublicRoute>
          }
        />
        <Route
          path="/register"
          element={
            <PublicRoute>
              <Register />
            </PublicRoute>
          }
        />

        {/* Protected Routes */}
        <Route
          path="/"
          element={
            <ProtectedRoute>
              <Layout />
            </ProtectedRoute>
          }
        >
          <Route index element={<Navigate to="/generate" replace />} />
          <Route path="generate" element={<Generate />} />
          <Route path="gallery" element={<Gallery />} />
          
          {/* Admin Routes */}
          <Route
            path="admin/credits"
            element={
              <ProtectedRoute requireAdmin>
                <Credits />
              </ProtectedRoute>
            }
          />
          <Route
            path="admin/members"
            element={
              <ProtectedRoute requireAdmin>
                <Members />
              </ProtectedRoute>
            }
          />
          <Route
            path="admin/providers"
            element={
              <ProtectedRoute requireAdmin>
                <Providers />
              </ProtectedRoute>
            }
          />
        </Route>

        {/* Fallback */}
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;

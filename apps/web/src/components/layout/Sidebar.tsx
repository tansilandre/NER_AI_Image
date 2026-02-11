import { NavLink, useNavigate } from 'react-router-dom';
import { 
  Sparkles, 
  Image, 
  Users, 
  Settings, 
  CreditCard, 
  LogOut,
  Hexagon
} from 'lucide-react';
import { useAuthStore } from '../../stores';

export function Sidebar() {
  const { user, organization, logout } = useAuthStore();
  const navigate = useNavigate();
  const isAdmin = user?.role === 'admin';

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <aside className="sidebar w-64 h-screen flex flex-col fixed left-0 top-0 z-50">
      {/* Logo */}
      <div className="p-6 border-b border-gray-200">
        <div className="flex items-center gap-2">
          <div className="w-8 h-8 bg-[var(--color-yellow)] flex items-center justify-center">
            <Hexagon className="w-5 h-5 text-black" />
          </div>
          <div>
            <h1 className="font-heading font-bold text-lg leading-tight">NER Studio</h1>
            <p className="text-xs text-gray-500">AI Image Generation</p>
          </div>
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex-1 py-4 px-3 overflow-y-auto">
        <div className="space-y-1">
          <NavLink
            to="/generate"
            className={({ isActive }) =>
              `flex items-center gap-3 px-3 py-2.5 text-sm font-medium transition-colors ${
                isActive
                  ? 'bg-[var(--color-yellow)]/10 text-black border-l-2 border-[var(--color-yellow)]'
                  : 'text-gray-600 hover:bg-gray-100'
              }`
            }
          >
            <Sparkles className="w-4 h-4" />
            Generate
          </NavLink>
          
          <NavLink
            to="/gallery"
            className={({ isActive }) =>
              `flex items-center gap-3 px-3 py-2.5 text-sm font-medium transition-colors ${
                isActive
                  ? 'bg-[var(--color-yellow)]/10 text-black border-l-2 border-[var(--color-yellow)]'
                  : 'text-gray-600 hover:bg-gray-100'
              }`
            }
          >
            <Image className="w-4 h-4" />
            Gallery
          </NavLink>
        </div>

        {isAdmin && (
          <>
            <div className="mt-8 mb-2 px-3">
              <p className="text-xs font-heading font-bold text-gray-400 uppercase tracking-wider">
                Admin
              </p>
            </div>
            
            <div className="space-y-1">
              <NavLink
                to="/admin/credits"
                className={({ isActive }) =>
                  `flex items-center gap-3 px-3 py-2.5 text-sm font-medium transition-colors ${
                    isActive
                      ? 'bg-[var(--color-yellow)]/10 text-black border-l-2 border-[var(--color-yellow)]'
                      : 'text-gray-600 hover:bg-gray-100'
                  }`
                }
              >
                <CreditCard className="w-4 h-4" />
                Credits
              </NavLink>
              
              <NavLink
                to="/admin/members"
                className={({ isActive }) =>
                  `flex items-center gap-3 px-3 py-2.5 text-sm font-medium transition-colors ${
                    isActive
                      ? 'bg-[var(--color-yellow)]/10 text-black border-l-2 border-[var(--color-yellow)]'
                      : 'text-gray-600 hover:bg-gray-100'
                  }`
                }
              >
                <Users className="w-4 h-4" />
                Members
              </NavLink>
              
              <NavLink
                to="/admin/providers"
                className={({ isActive }) =>
                  `flex items-center gap-3 px-3 py-2.5 text-sm font-medium transition-colors ${
                    isActive
                      ? 'bg-[var(--color-yellow)]/10 text-black border-l-2 border-[var(--color-yellow)]'
                      : 'text-gray-600 hover:bg-gray-100'
                  }`
                }
              >
                <Settings className="w-4 h-4" />
                Providers
              </NavLink>
            </div>
          </>
        )}
      </nav>

      {/* User */}
      <div className="p-4 border-t border-gray-200">
        <div className="flex items-center gap-3 mb-3">
          <div className="w-8 h-8 rounded-full bg-[var(--color-blue)] flex items-center justify-center text-white text-sm font-medium">
            {user?.name?.[0] || 'U'}
          </div>
          <div className="flex-1 min-w-0">
            <p className="text-sm font-medium truncate">{user?.name}</p>
            <p className="text-xs text-gray-500 truncate">{organization?.name}</p>
          </div>
        </div>
        
        <button
          onClick={handleLogout}
          className="flex items-center gap-2 text-sm text-gray-600 hover:text-red-600 transition-colors"
        >
          <LogOut className="w-4 h-4" />
          Logout
        </button>
      </div>
    </aside>
  );
}

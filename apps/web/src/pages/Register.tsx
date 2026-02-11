import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Hexagon } from 'lucide-react';
import { Button, Input, Card, CardHeader } from '../components/ui';
import { useAuthStore } from '../stores';
import { toast } from 'sonner';

export function Register() {
  const [formData, setFormData] = useState({
    email: '',
    password: '',
    confirmPassword: '',
    fullName: '',
    orgName: '',
  });
  const [isLoading, setIsLoading] = useState(false);
  const { register } = useAuthStore();
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (formData.password !== formData.confirmPassword) {
      toast.error('Passwords do not match');
      return;
    }

    if (formData.password.length < 8) {
      toast.error('Password must be at least 8 characters');
      return;
    }

    setIsLoading(true);
    
    try {
      await register({
        email: formData.email,
        password: formData.password,
        full_name: formData.fullName,
        org_name: formData.orgName,
      });
      toast.success('Account created successfully!');
      navigate('/generate');
    } catch (error: any) {
      toast.error(error.response?.data?.message || 'Failed to create account');
    } finally {
      setIsLoading(false);
    }
  };

  const updateField = (field: string, value: string) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-[var(--color-off-white)] p-4">
      <div className="w-full max-w-md">
        {/* Logo */}
        <div className="flex items-center justify-center gap-2 mb-8">
          <div className="w-12 h-12 bg-[var(--color-yellow)] flex items-center justify-center">
            <Hexagon className="w-7 h-7 text-black" />
          </div>
          <div>
            <h1 className="font-heading font-bold text-2xl">NER Studio</h1>
            <p className="text-xs text-gray-500">AI Image Generation</p>
          </div>
        </div>

        <Card className="border-t-4 border-t-[var(--color-yellow)]">
          <CardHeader 
            title="Create Account"
            subtitle="Get started with your organization today."
          />
          
          <form onSubmit={handleSubmit} className="space-y-4">
            <Input
              label="Full Name"
              placeholder="John Doe"
              value={formData.fullName}
              onChange={(e) => updateField('fullName', e.target.value)}
              required
            />
            
            <Input
              label="Organization Name"
              placeholder="Acme Inc."
              value={formData.orgName}
              onChange={(e) => updateField('orgName', e.target.value)}
              required
            />
            
            <Input
              label="Email"
              type="email"
              placeholder="name@company.com"
              value={formData.email}
              onChange={(e) => updateField('email', e.target.value)}
              required
            />
            
            <Input
              label="Password"
              type="password"
              placeholder="••••••••"
              value={formData.password}
              onChange={(e) => updateField('password', e.target.value)}
              required
            />
            
            <Input
              label="Confirm Password"
              type="password"
              placeholder="••••••••"
              value={formData.confirmPassword}
              onChange={(e) => updateField('confirmPassword', e.target.value)}
              required
            />
            
            <Button
              type="submit"
              variant="primary"
              size="lg"
              isLoading={isLoading}
              className="w-full"
            >
              Create Account
            </Button>
          </form>
          
          <p className="mt-6 text-center text-sm text-gray-600">
            Already have an account?{' '}
            <Link to="/login" className="font-medium text-[var(--color-blue)] hover:underline">
              Sign in
            </Link>
          </p>
        </Card>
      </div>
    </div>
  );
}

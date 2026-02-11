import { useEffect, useState } from 'react';
import { UserPlus, Mail } from 'lucide-react';
import { Card, CardHeader, Button, Badge } from '../../components/ui';
import { adminApi } from '../../lib/api';
import { toast } from 'sonner';

interface Member {
  id: string;
  email: string;
  name: string;
  role: string;
  created_at: string;
  credits_spent: number;
}

export function Members() {
  const [members, setMembers] = useState<Member[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [showInviteModal, setShowInviteModal] = useState(false);
  const [inviteEmail, setInviteEmail] = useState('');

  const fetchMembers = async () => {
    try {
      setIsLoading(true);
      const response = await adminApi.listMembers();
      setMembers(response.data.members || []);
    } catch (error) {
      toast.error('Failed to fetch members');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchMembers();
  }, []);

  const handleInvite = async () => {
    if (!inviteEmail) {
      toast.error('Please enter an email address');
      return;
    }

    try {
      await adminApi.inviteMember(inviteEmail);
      toast.success('Invitation sent');
      setShowInviteModal(false);
      setInviteEmail('');
      fetchMembers();
    } catch (error: any) {
      toast.error(error.response?.data?.message || 'Failed to send invitation');
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="font-heading font-bold text-2xl">Members</h1>
          <p className="text-gray-600">Manage organization members</p>
        </div>
        
        <Button onClick={() => setShowInviteModal(true)} variant="primary">
          <UserPlus className="w-4 h-4" />
          Invite Member
        </Button>
      </div>

      {/* Members List */}
      <Card>
        <CardHeader title="Team Members" />
        
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-200">
                <th className="text-left py-3 px-4 font-heading font-semibold text-sm">
                  Member
                </th>
                <th className="text-left py-3 px-4 font-heading font-semibold text-sm">
                  Role
                </th>
                <th className="text-left py-3 px-4 font-heading font-semibold text-sm">
                  Credits Used
                </th>
                <th className="text-left py-3 px-4 font-heading font-semibold text-sm">
                  Joined
                </th>
              </tr>
            </thead>
            <tbody>
              {members.length === 0 ? (
                <tr>
                  <td colSpan={4} className="py-8 text-center text-gray-500">
                    No members yet
                  </td>
                </tr>
              ) : (
                members.map((member) => (
                  <tr key={member.id} className="border-b border-gray-100">
                    <td className="py-3 px-4">
                      <div className="flex items-center gap-3">
                        <div className="w-8 h-8 rounded-full bg-[var(--color-blue)] flex items-center justify-center text-white text-sm font-medium">
                          {member.name[0]}
                        </div>
                        <div>
                          <p className="font-medium">{member.name}</p>
                          <p className="text-xs text-gray-500">{member.email}</p>
                        </div>
                      </div>
                    </td>
                    <td className="py-3 px-4">
                      <Badge variant={member.role === 'admin' ? 'primary' : 'default'}>
                        {member.role}
                      </Badge>
                    </td>
                    <td className="py-3 px-4 text-sm">{member.credits_spent || 0}</td>
                    <td className="py-3 px-4 text-sm text-gray-600">
                      {new Date(member.created_at).toLocaleDateString()}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </Card>

      {/* Invite Modal */}
      {showInviteModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <Card className="w-full max-w-md mx-4">
            <CardHeader title="Invite Member" subtitle="Invite a new member to your organization" />
            
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Email Address
                </label>
                <div className="relative">
                  <Mail className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
                  <input
                    type="email"
                    value={inviteEmail}
                    onChange={(e) => setInviteEmail(e.target.value)}
                    className="input w-full pl-10 pr-4 py-3 border-2 border-gray-200 focus:border-[var(--color-yellow)] focus:outline-none"
                    placeholder="member@company.com"
                  />
                </div>
              </div>
              
              <div className="flex gap-3 pt-4">
                <Button
                  variant="outline"
                  className="flex-1"
                  onClick={() => setShowInviteModal(false)}
                >
                  Cancel
                </Button>
                <Button
                  variant="primary"
                  className="flex-1"
                  onClick={handleInvite}
                  isLoading={isLoading}
                >
                  <Mail className="w-4 h-4" />
                  Send Invite
                </Button>
              </div>
            </div>
          </Card>
        </div>
      )}
    </div>
  );
}

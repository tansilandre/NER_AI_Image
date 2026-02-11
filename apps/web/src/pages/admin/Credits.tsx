import { useEffect, useState } from 'react';
import { Plus, History } from 'lucide-react';
import { Card, CardHeader, Button, Badge } from '../../components/ui';
import { adminApi } from '../../lib/api';
import { useAuthStore } from '../../stores';
import { toast } from 'sonner';

interface CreditTransaction {
  id: string;
  amount: number;
  balance: number;
  note: string;
  created_at: string;
  user_name: string;
}

export function Credits() {
  const { organization } = useAuthStore();
  const [transactions, setTransactions] = useState<CreditTransaction[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [showAddModal, setShowAddModal] = useState(false);
  const [amount, setAmount] = useState('');
  const [note, setNote] = useState('');

  const fetchTransactions = async () => {
    try {
      setIsLoading(true);
      const response = await adminApi.listTransactions();
      setTransactions(response.data.transactions || []);
    } catch (error) {
      toast.error('Failed to fetch transactions');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchTransactions();
  }, []);

  const handleAddCredits = async () => {
    if (!amount || Number(amount) <= 0) {
      toast.error('Please enter a valid amount');
      return;
    }

    try {
      await adminApi.addCredits(Number(amount), note);
      toast.success('Credits added successfully');
      setShowAddModal(false);
      setAmount('');
      setNote('');
      fetchTransactions();
    } catch (error: any) {
      toast.error(error.response?.data?.message || 'Failed to add credits');
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="font-heading font-bold text-2xl">Credits</h1>
          <p className="text-gray-600">Manage organization credits</p>
        </div>
        
        <Button onClick={() => setShowAddModal(true)} variant="primary">
          <Plus className="w-4 h-4" />
          Add Credits
        </Button>
      </div>

      {/* Balance Card */}
      <Card variant="accent">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-sm text-gray-500">Current Balance</p>
            <p className="font-heading font-bold text-4xl mt-1">
              {organization?.credits || 0}
            </p>
            <p className="text-sm text-gray-500 mt-2">
              Credits available for image generation
            </p>
          </div>
          
          <div className="w-16 h-16 bg-[var(--color-yellow)] flex items-center justify-center">
            <History className="w-8 h-8 text-black" />
          </div>
        </div>
      </Card>

      {/* Transactions */}
      <Card>
        <CardHeader title="Transaction History" />
        
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-200">
                <th className="text-left py-3 px-4 font-heading font-semibold text-sm">
                  Date
                </th>
                <th className="text-left py-3 px-4 font-heading font-semibold text-sm">
                  User
                </th>
                <th className="text-left py-3 px-4 font-heading font-semibold text-sm">
                  Amount
                </th>
                <th className="text-left py-3 px-4 font-heading font-semibold text-sm">
                  Balance
                </th>
                <th className="text-left py-3 px-4 font-heading font-semibold text-sm">
                  Note
                </th>
              </tr>
            </thead>
            <tbody>
              {transactions.length === 0 ? (
                <tr>
                  <td colSpan={5} className="py-8 text-center text-gray-500">
                    No transactions yet
                  </td>
                </tr>
              ) : (
                transactions.map((tx) => (
                  <tr key={tx.id} className="border-b border-gray-100">
                    <td className="py-3 px-4 text-sm">
                      {new Date(tx.created_at).toLocaleString()}
                    </td>
                    <td className="py-3 px-4 text-sm">{tx.user_name}</td>
                    <td className="py-3 px-4 text-sm">
                      <Badge variant={tx.amount > 0 ? 'success' : 'error'}>
                        {tx.amount > 0 ? '+' : ''}{tx.amount}
                      </Badge>
                    </td>
                    <td className="py-3 px-4 text-sm font-medium">{tx.balance}</td>
                    <td className="py-3 px-4 text-sm text-gray-600">{tx.note || '-'}</td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </Card>

      {/* Add Credits Modal */}
      {showAddModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <Card className="w-full max-w-md mx-4">
            <CardHeader title="Add Credits" subtitle="Add credits to organization" />
            
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Amount
                </label>
                <input
                  type="number"
                  value={amount}
                  onChange={(e) => setAmount(e.target.value)}
                  className="input w-full px-4 py-3 border-2 border-gray-200 focus:border-[var(--color-yellow)] focus:outline-none"
                  placeholder="Enter amount"
                />
              </div>
              
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Note (optional)
                </label>
                <input
                  type="text"
                  value={note}
                  onChange={(e) => setNote(e.target.value)}
                  className="input w-full px-4 py-3 border-2 border-gray-200 focus:border-[var(--color-yellow)] focus:outline-none"
                  placeholder="e.g., Monthly top-up"
                />
              </div>
              
              <div className="flex gap-3 pt-4">
                <Button
                  variant="outline"
                  className="flex-1"
                  onClick={() => setShowAddModal(false)}
                >
                  Cancel
                </Button>
                <Button
                  variant="primary"
                  className="flex-1"
                  onClick={handleAddCredits}
                  isLoading={isLoading}
                >
                  Add Credits
                </Button>
              </div>
            </div>
          </Card>
        </div>
      )}
    </div>
  );
}

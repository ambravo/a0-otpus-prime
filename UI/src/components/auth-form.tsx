import { useState, useEffect } from 'react';
import { toast } from 'sonner';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Shield, Key, Globe, Terminal } from 'lucide-react';

declare global {
  interface Window {
    formData?: {
      chatId: string;
      messageId: string;
      signature: string;
      authType: string;
      csrfToken: string;
    };
  }
}

interface FormData {
  domain?: string;
  access_token?: string;
  client_id?: string;
  client_secret?: string;
}

// Initialize default form data for development
if (typeof window !== 'undefined' && !window.formData) {
  window.formData = {
    chatId: "8109655141",
    messageId: "69",
    signature: "888831bc4e853c853b23bbb901fec502f28144bbeca43b785d813b11b249cdce",
    authType: "tenant_personal",
    csrfToken: "fZWveaAKp3Qc2PG5DmOQW3BuRxZVfstO"
  };
}

export default function AuthForm() {
  const [formData, setFormData] = useState<FormData>({});
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [authType, setAuthType] = useState<string>('auth_client_credentials');

  useEffect(() => {
    if (window.formData?.authType) {
      setAuthType(window.formData.authType);
    }
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    if (!window.formData) {
      setError('Missing configuration data');
      setLoading(false);
      return;
    }

    try {
      const response = await fetch('/bot/auth-form', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': window.formData.csrfToken,
        },
        body: JSON.stringify({
          ...formData,
          chat_id: window.formData.chatId,
          message_id: window.formData.messageId,
          signature: window.formData.signature,
          auth_type: window.formData.authType,
        }),
      });

      const data = await response.json();
      if (!response.ok) throw new Error(data.error || 'Authentication failed');

      toast.success('Authentication successful!');
      window.close();
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'An error occurred';
      setError(errorMessage);
      toast.error('Authentication failed');
    } finally {
      setLoading(false);
    }
  };

  const renderForm = () => {
    switch (authType) {
      case 'tenant_personal':
        return (
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="domain">Domain</Label>
              <div className="relative">
                <Globe className="absolute left-3 top-2.5 h-5 w-5 text-muted-foreground" />
                <Input
                  id="domain"
                  placeholder="your-tenant.auth0.com"
                  className="pl-10"
                  onChange={(e) =>
                    setFormData({ ...formData, domain: e.target.value })
                  }
                />
              </div>
            </div>
          </div>
        );

      case 'auth_ephemeral':
        return (
          <div className="space-y-10">
            <div className="space-y-2">
              <Label htmlFor="access_token">Access Token</Label>
              <div className="relative">
                <Key className="absolute left-3 top-2.5 h-5 w-5 text-muted-foreground" />
                <Input
                  id="access_token"
                  type="password"
                  placeholder="Access Token"
                  className="pl-10"
                  onChange={(e) =>
                    setFormData({ ...formData, access_token: e.target.value })
                  }
                />
              </div>
            </div>
          </div>
        );

      case 'auth_client_credentials':
        return (
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="domain">Domain</Label>
              <div className="relative">
                <Globe className="absolute left-3 top-2.5 h-5 w-5 text-muted-foreground" />
                <Input
                  id="domain"
                  placeholder="your-tenant.auth0.com"
                  className="pl-10"
                  onChange={(e) =>
                    setFormData({ ...formData, domain: e.target.value })
                  }
                />
              </div>
            </div>
            <div className="space-y-2">
              <Label htmlFor="client_id">Client ID</Label>
              <div className="relative">
                <Shield className="absolute left-3 top-2.5 h-5 w-5 text-muted-foreground" />
                <Input
                  id="client_id"
                  placeholder="Client ID"
                  className="pl-10"
                  onChange={(e) =>
                    setFormData({ ...formData, client_id: e.target.value })
                  }
                />
              </div>
            </div>
            <div className="space-y-2">
              <Label htmlFor="client_secret">Client Secret</Label>
              <div className="relative">
                <Key className="absolute left-3 top-2.5 h-5 w-5 text-muted-foreground" />
                <Input
                  id="client_secret"
                  type="password"
                  placeholder="Client Secret"
                  className="pl-10"
                  onChange={(e) =>
                    setFormData({ ...formData, client_secret: e.target.value })
                  }
                />
              </div>
            </div>
          </div>
        );

      default:
        return (
          <Alert variant="destructive">
            <AlertDescription>Invalid authentication type</AlertDescription>
          </Alert>
        );
    }
  };

  if (!window.formData) {
    return (
      <Alert variant="destructive">
        <AlertDescription>Missing configuration data</AlertDescription>
      </Alert>
    );
  }

  return (
    <div className="w-full max-w-lg mx-auto px-10">
      <Card>
        <CardHeader>
          <CardTitle className="text-2xl font-bold text-center">
            Bind your Auth0 Tenant
          </CardTitle>
        </CardHeader>
        <CardContent>
          {error && (
            <Alert variant="destructive" className="mb-6">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
          <form onSubmit={handleSubmit} className="space-y-6">
            {renderForm()}
            <Button
              type="submit"
              className="w-full"
              disabled={loading}
              size="lg"
            >
              {loading ? 'Authenticating...' : 'Submit'}
            </Button>
          </form>
          <div className="space-y-12">
            <div/>
              <Alert>
                <Terminal className="h-6 w-6" />
                <AlertTitle className="text-lg"><b>Where do I get this Information?</b></AlertTitle>
                <AlertDescription>
                <br/>
                <i><u>Access Token:</u> </i><br/>On your Auth0 Dashboard, navigate to Applications &gt; APIs &gt; Auth0 Management API. <br/>Select the API Explorer tab and locate an auto-generated token in the Token section.
                <br/>
                <br/>
                <i><u>Client ID & Secret:</u></i><br/> On your Auth0 Dashboard, navigate to Applications &gt; Aplications.<br/>Select an Application that can leverage the Management API. For instance "Auth0 Dashboard Backend Management Client".
                <br/>
                <br/>
                <i><u>Domain:</u></i><br/> On your Auth0 Dashboard, navigate to Settings &gt; Custom Domains.
                </AlertDescription>
              </Alert>
            </div>

        </CardContent>
      </Card>
    </div>
  );
}
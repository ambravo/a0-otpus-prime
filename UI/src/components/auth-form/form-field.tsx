import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Shield, Key, Globe } from 'lucide-react';

interface FormFieldProps {
  id: string;
  label: string;
  type?: string;
  placeholder: string;
  icon: typeof Shield| typeof Key| typeof Globe ;
  onChange: (value: string) => void;
}

export function FormField({
  id,
  label,
  type = 'text',
  placeholder,
  icon: Icon,
  onChange,
}: FormFieldProps) {
  return (
    <div className="space-y-2">
      <Label htmlFor={id}>{label}</Label>
      <div className="relative">
        <Icon className="absolute left-3 top-2.5 h-5 w-5 text-muted-foreground" />
        <Input
          id={id}
          type={type}
          placeholder={placeholder}
          className="pl-10"
          onChange={(e) => onChange(e.target.value)}
        />
      </div>
    </div>
  );
}
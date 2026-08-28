import type { ComponentType } from 'svelte';
import { Fingerprint, GitBranch, KeyRound, Monitor, Palette, Settings, Shield } from 'lucide-svelte';

export type SettingsNavItem = {
	label: string;
	href: string;
	icon: ComponentType;
	exact?: boolean;
};

export type SettingsNavGroup = {
	label: string;
	items: SettingsNavItem[];
};

export const SETTINGS_NAV: SettingsNavGroup[] = [
	{
		label: 'Panel',
		items: [
			{
				label: 'Branding',
				href: '/dashboard/management/settings',
				icon: Settings,
				exact: true
			},
			{ label: 'Authentication', href: '/dashboard/management/settings/auth', icon: Shield },
			{ label: 'Appearance', href: '/dashboard/management/settings/appearance', icon: Palette }
		]
	},
	{
		label: 'Access',
		items: [
			{ label: 'API tokens', href: '/dashboard/management/settings/tokens', icon: KeyRound },
			{ label: 'Node keys', href: '/dashboard/management/settings/keys', icon: Fingerprint }
		]
	},
	{
		label: 'Subscription',
		items: [
			{
				label: 'Subscription',
				href: '/dashboard/management/settings/subscription',
				icon: Monitor
			},
			{
				label: 'Response rules',
				href: '/dashboard/management/settings/response-rules',
				icon: GitBranch
			}
		]
	}
];

<script lang="ts">
	import { Monitor, Moon, Sun } from 'lucide-svelte';
	import * as Select from '$lib/components/ui/select';
	import * as ToggleGroup from '$lib/components/ui/toggle-group';
	import { Switch } from '$lib/components/ui/switch';
	import SettingsCard from '$lib/components/remnawave/SettingsCard.svelte';
	import SettingsRow from '$lib/components/remnawave/SettingsRow.svelte';
	import { pageChrome } from '$lib/features/layout/page-chrome.svelte';
	import { preferencesState } from '$lib/features/preferences/preferences.state.svelte';

	const fontSizeLabels: Record<string, string> = {
		small: 'Small',
		default: 'Default',
		large: 'Large'
	};
	const lightThemeLabels: Record<string, string> = {
		light: 'Light',
		'rose-light': 'Rose',
		'blue-light': 'Blue'
	};
	const darkThemeLabels: Record<string, string> = {
		dark: 'Dark',
		'dark-gray': 'Gray',
		'amethyst-dark': 'Amethyst',
		'emerald-dark': 'Emerald',
		'cyber-77': 'Cyber',
		'blade-49': 'Blade',
		pipboy: 'Pip-Boy'
	};

	$effect.pre(() => {
		const token = pageChrome.set({ title: 'Appearance' });
		return () => pageChrome.clear(token);
	});
</script>

<h1 class="text-2xl font-semibold text-[var(--color-text-primary)]">Appearance</h1>
<p class="mt-1 text-sm text-[var(--color-text-tertiary)]">Local to this browser.</p>

<SettingsCard title="Interface">
	<SettingsRow title="Font size" description="Scales the whole panel.">
		<Select.Root
			type="single"
			value={preferencesState.fontSize}
			onValueChange={(value) => value && preferencesState.setFontSize(value as 'small' | 'default' | 'large')}
		>
			<Select.Trigger size="sm" class="w-[130px]">{fontSizeLabels[preferencesState.fontSize]}</Select.Trigger>
			<Select.Content>
				<Select.Item value="small">Small</Select.Item>
				<Select.Item value="default">Default</Select.Item>
				<Select.Item value="large">Large</Select.Item>
			</Select.Content>
		</Select.Root>
	</SettingsRow>
	<SettingsRow title="Pointer cursors" description="Show pointer on clickable rows.">
		<Switch
			size="sm"
			checked={preferencesState.pointerCursors}
			onCheckedChange={(value) => preferencesState.setPointerCursors(value)}
		/>
	</SettingsRow>
	<SettingsRow title="Theme mode">
		<ToggleGroup.Root
			type="single"
			variant="outline"
			size="sm"
			value={preferencesState.themeMode}
			onValueChange={(value) => value && preferencesState.setThemeMode(value as 'system' | 'light' | 'dark')}
		>
			<ToggleGroup.Item value="system" aria-label="System"><Monitor size={14} /></ToggleGroup.Item>
			<ToggleGroup.Item value="light" aria-label="Light"><Sun size={14} /></ToggleGroup.Item>
			<ToggleGroup.Item value="dark" aria-label="Dark"><Moon size={14} /></ToggleGroup.Item>
		</ToggleGroup.Root>
	</SettingsRow>
	<SettingsRow title="Light theme">
		<Select.Root
			type="single"
			value={preferencesState.lightTheme}
			onValueChange={(value) =>
				value && preferencesState.setLightTheme(value as 'light' | 'rose-light' | 'blue-light')}
		>
			<Select.Trigger size="sm" class="w-[130px]">{lightThemeLabels[preferencesState.lightTheme]}</Select.Trigger>
			<Select.Content>
				{#each Object.entries(lightThemeLabels) as [value, label]}
					<Select.Item {value}>{label}</Select.Item>
				{/each}
			</Select.Content>
		</Select.Root>
	</SettingsRow>
	<SettingsRow title="Dark theme">
		<Select.Root
			type="single"
			value={preferencesState.darkTheme}
			onValueChange={(value) =>
				value &&
				preferencesState.setDarkTheme(
					value as 'dark' | 'dark-gray' | 'amethyst-dark' | 'emerald-dark' | 'cyber-77' | 'blade-49' | 'pipboy'
				)}
		>
			<Select.Trigger size="sm" class="w-[130px]">{darkThemeLabels[preferencesState.darkTheme]}</Select.Trigger>
			<Select.Content>
				{#each Object.entries(darkThemeLabels) as [value, label]}
					<Select.Item {value}>{label}</Select.Item>
				{/each}
			</Select.Content>
		</Select.Root>
	</SettingsRow>
</SettingsCard>

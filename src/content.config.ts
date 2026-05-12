import { defineCollection, z } from 'astro:content';
import { glob } from 'astro/loaders';
import { STAGE_IDS, type Stage } from './lib/stages';

const command = z.object({
	run: z.string(),
	description: z.string().optional(),
});

const upstream = z.object({
	name: z.string(),
	url: z.string().url(),
});

const supportedDevice = z.enum(['deck', 'steam-machine', 'any']);

const commands = defineCollection({
	loader: glob({
		pattern: '**/*.{md,mdx}',
		base: './src/content/commands',
	}),
	schema: z.object({
		title: z.string(),
		category: z.enum(STAGE_IDS as unknown as readonly [Stage, ...Stage[]]),
		order: z.number().default(100),
		summary: z.string(),
		commands: z.array(command).min(1),
		idempotent: z.boolean().default(true),
		reversible: z.boolean().default(false),
		undo: z.string().optional(),
		upstream: upstream.optional(),
		supported_devices: z.array(supportedDevice).default(['any']),
		deck_only: z.boolean().default(false),
		danger: z.enum(['low', 'medium', 'high']).optional(),
		tags: z.array(z.string()).default([]),
		icon: z.string().optional(),
		group: z.string().optional(),
	}),
});

const spellbook = defineCollection({
	loader: glob({
		pattern: '**/*.{md,mdx}',
		base: './src/content/spellbook',
	}),
	schema: z.object({
		title: z.string(),
		teaser: z.string(),
		icon: z.string(),
		status: z.enum(['planned', 'research', 'draft']),
		order: z.number().default(100),
	}),
});

export const collections = { commands, spellbook };

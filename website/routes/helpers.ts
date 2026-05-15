// routes/helpers.ts — Shared route handler utilities

import type { Context } from "@oak/oak";

/**
 * Extracts user data from ctx.state.user and formats it for template rendering.
 * Returns the user object with an `isAdmin` flag, or null if not authenticated.
 */
export function getTemplateUser(ctx: Context): {
  id: number;
  name: string;
  email: string;
  roles: Array<{ id: number; name: string; description: string }>;
  isAdmin: boolean;
} | null {
  const user = ctx.state.user as {
    id: number;
    name: string;
    email: string;
    roles: Array<{ id: number; name: string; description: string }>;
  } | null | undefined;

  if (!user) return null;

  const isAdmin = user.roles.some((role) => role.name === "admin");

  return {
    ...user,
    isAdmin,
  };
}

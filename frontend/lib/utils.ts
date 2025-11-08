import { type ClassValue, clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function formatDate(date: string | Date): string {
  return new Date(date).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  })
}

export function formatDateTime(date: string | Date): string {
  return new Date(date).toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
  })
}

export function formatNumber(num: number, decimals: number = 1): string {
  return num.toFixed(decimals)
}

export function getTeamColor(team: string): string {
  const colors: Record<string, string> = {
    KC: '#E31837',
    BUF: '#00338D',
    SF: '#AA0000',
    MIN: '#4F2683',
    // Add more teams as needed
  }
  return colors[team] || '#1e40af'
}

export function getPositionColor(position: string): string {
  const colors: Record<string, string> = {
    QB: '#3b82f6',
    RB: '#10b981',
    WR: '#f59e0b',
    TE: '#8b5cf6',
    K: '#64748b',
    DEF: '#dc2626',
  }
  return colors[position] || '#6b7280'
}


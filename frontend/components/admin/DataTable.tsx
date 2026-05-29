'use client';

import { ReactNode } from 'react';

export interface Column<T> {
  key: string;
  header: string;
  render: (row: T) => ReactNode;
  width?: number | string;
}

interface DataTableProps<T> {
  columns: Column<T>[];
  rows: T[];
  rowKey: (row: T) => string | number;
  empty?: string;
  onRowClick?: (row: T) => void;
}

export function DataTable<T>({ columns, rows, rowKey, empty = 'No data', onRowClick }: DataTableProps<T>) {
  if (rows.length === 0) {
    return (
      <div style={{ padding: 40, textAlign: 'center', color: 'var(--text-tertiary)' }}>
        {empty}
      </div>
    );
  }

  return (
    <div style={{ overflowX: 'auto', borderRadius: 12, border: '1px solid var(--border)' }}>
      <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 14 }}>
        <thead>
          <tr style={{ background: 'var(--bg-subtle)', borderBottom: '1px solid var(--border-strong)' }}>
            {columns.map((c) => (
              <th
                key={c.key}
                style={{
                  textAlign: 'left',
                  padding: '12px 16px',
                  fontWeight: 600,
                  color: 'var(--text-secondary)',
                  width: c.width,
                  whiteSpace: 'nowrap',
                }}
              >
                {c.header}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map((row, i) => {
            const stagger = i < 8;
            return (
              <tr
                key={rowKey(row)}
                className={onRowClick ? 'admin-row admin-row-clickable' : 'admin-row'}
                onClick={onRowClick ? () => onRowClick(row) : undefined}
                style={{
                  borderBottom: '1px solid var(--border)',
                  animation: stagger ? `row-in 220ms ease-out backwards` : undefined,
                  animationDelay: stagger ? `${i * 18}ms` : undefined,
                }}
              >
                {columns.map((c) => (
                  <td key={c.key} style={{ padding: '12px 16px', color: 'var(--text)', verticalAlign: 'middle' }}>
                    {c.render(row)}
                  </td>
                ))}
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}

'use client';

import { ReactNode, useMemo, useState } from 'react';
import { ArrowDown, ArrowUp, ChevronsUpDown } from 'lucide-react';

export interface Column<T> {
  key: string;
  header: string;
  render: (row: T) => ReactNode;
  width?: number | string;
  align?: 'left' | 'right' | 'center';
  /** Provide to make the column sortable. Returns the value to sort by. */
  sortValue?: (row: T) => string | number | null | undefined;
}

interface DataTableProps<T> {
  columns: Column<T>[];
  rows: T[];
  rowKey: (row: T) => string | number;
  empty?: string;
  emptyIcon?: ReactNode;
  onRowClick?: (row: T) => void;
  loading?: boolean;
  skeletonRows?: number;
}

type SortState = { key: string; dir: 'asc' | 'desc' } | null;

export function DataTable<T>({
  columns,
  rows,
  rowKey,
  empty = 'No data',
  emptyIcon,
  onRowClick,
  loading = false,
  skeletonRows = 6,
}: DataTableProps<T>) {
  const [sort, setSort] = useState<SortState>(null);

  const sortedRows = useMemo(() => {
    if (!sort) return rows;
    const col = columns.find((c) => c.key === sort.key);
    if (!col?.sortValue) return rows;
    const dir = sort.dir === 'asc' ? 1 : -1;
    return [...rows].sort((a, b) => {
      const av = col.sortValue!(a);
      const bv = col.sortValue!(b);
      if (av == null && bv == null) return 0;
      if (av == null) return 1;
      if (bv == null) return -1;
      if (typeof av === 'number' && typeof bv === 'number') return (av - bv) * dir;
      return String(av).localeCompare(String(bv)) * dir;
    });
  }, [rows, sort, columns]);

  const toggleSort = (key: string) => {
    setSort((prev) => {
      if (prev?.key !== key) return { key, dir: 'asc' };
      if (prev.dir === 'asc') return { key, dir: 'desc' };
      return null;
    });
  };

  const tableShell = (body: ReactNode) => (
    <div style={{ overflowX: 'auto', borderRadius: 12, border: '1px solid var(--border)' }}>
      <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 14 }}>
        <thead>
          <tr style={{ background: 'var(--bg-subtle)', borderBottom: '1px solid var(--border-strong)' }}>
            {columns.map((c) => {
              const sortable = !!c.sortValue && !loading;
              const active = sort?.key === c.key;
              return (
                <th
                  key={c.key}
                  onClick={sortable ? () => toggleSort(c.key) : undefined}
                  className={sortable ? 'admin-th-sortable' : undefined}
                  aria-sort={active ? (sort!.dir === 'asc' ? 'ascending' : 'descending') : undefined}
                  style={{
                    textAlign: c.align ?? 'left',
                    padding: '12px 16px',
                    fontWeight: 600,
                    color: active ? 'var(--text)' : 'var(--text-secondary)',
                    width: c.width,
                    whiteSpace: 'nowrap',
                    cursor: sortable ? 'pointer' : undefined,
                    userSelect: 'none',
                  }}
                >
                  <span
                    style={{
                      display: 'inline-flex',
                      alignItems: 'center',
                      gap: 5,
                      justifyContent: c.align === 'right' ? 'flex-end' : 'flex-start',
                    }}
                  >
                    {c.header}
                    {sortable &&
                      (active ? (
                        sort!.dir === 'asc' ? (
                          <ArrowUp size={13} strokeWidth={2.4} />
                        ) : (
                          <ArrowDown size={13} strokeWidth={2.4} />
                        )
                      ) : (
                        <ChevronsUpDown size={13} strokeWidth={1.8} style={{ opacity: 0.4 }} />
                      ))}
                  </span>
                </th>
              );
            })}
          </tr>
        </thead>
        <tbody>{body}</tbody>
      </table>
    </div>
  );

  if (loading) {
    return tableShell(
      Array.from({ length: skeletonRows }).map((_, i) => (
        <tr key={`sk-${i}`} style={{ borderBottom: '1px solid var(--border)' }}>
          {columns.map((c) => (
            <td key={c.key} style={{ padding: '12px 16px' }}>
              <div
                className="admin-skeleton"
                style={{ height: 14, width: `${55 + ((i * 7 + c.key.length * 11) % 35)}%`, borderRadius: 6 }}
              />
            </td>
          ))}
        </tr>
      )),
    );
  }

  if (sortedRows.length === 0) {
    return (
      <div
        style={{
          padding: '48px 24px',
          textAlign: 'center',
          color: 'var(--text-tertiary)',
          border: '1px dashed var(--border)',
          borderRadius: 12,
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          gap: 10,
        }}
      >
        {emptyIcon}
        <span style={{ fontSize: 14 }}>{empty}</span>
      </div>
    );
  }

  return tableShell(
    sortedRows.map((row, i) => {
      const stagger = i < 8 && !sort;
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
            <td
              key={c.key}
              style={{ padding: '12px 16px', color: 'var(--text)', verticalAlign: 'middle', textAlign: c.align ?? 'left' }}
            >
              {c.render(row)}
            </td>
          ))}
        </tr>
      );
    }),
  );
}

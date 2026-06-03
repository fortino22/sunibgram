'use client';

import { useQuery } from '@tanstack/react-query';
import { getFacultyLocations, type FacultyLocation } from '@/lib/services/facultyLocations';
import { useAuthStore } from '@/stores/authStore';

// Faculty locations are near-static reference data — cache aggressively.
const FACULTY_LOCATION_STALE = 24 * 60 * 60 * 1000; // 24h
const FACULTY_LOCATION_GC = 7 * 24 * 60 * 60 * 1000; // 7d

// useFacultyLocations fetches the campus list. includeAll=true (event audience)
// adds the "All Regions" option; the default (register/profile) omits it.
export function useFacultyLocations(includeAll = false) {
  const token = useAuthStore((s) => s.token);
  return useQuery<FacultyLocation[]>({
    queryKey: ['faculty-locations', includeAll],
    queryFn: () => getFacultyLocations(includeAll, token ?? undefined),
    staleTime: FACULTY_LOCATION_STALE,
    gcTime: FACULTY_LOCATION_GC,
  });
}

import { api } from '../api';

interface BackendFacultyLocation {
  id: number;
  name: string;
  address?: string;
  region: string;
  is_all_regions: boolean;
}

interface FacultyLocationListResponse {
  data: BackendFacultyLocation[];
}

export interface FacultyLocation {
  id: number;
  name: string;
  address: string;
  region: string;
  isAllRegions: boolean;
}

function mapFacultyLocation(raw: BackendFacultyLocation): FacultyLocation {
  return {
    id: raw.id,
    name: raw.name,
    address: raw.address ?? '',
    region: raw.region,
    isAllRegions: raw.is_all_regions,
  };
}

// Public endpoint — token optional (the signup form reads it pre-auth).
// includeAll=true adds the "All Regions" sentinel for the event-audience picker;
// the default (register/profile) omits it.
export async function getFacultyLocations(includeAll = false, token?: string): Promise<FacultyLocation[]> {
  const qs = includeAll ? '?include_all=true' : '';
  const res = await api.get<FacultyLocationListResponse>(`/api/faculty-locations${qs}`, token);
  return (res.data ?? []).map(mapFacultyLocation);
}

// groupByRegion preserves list order (already sorted by region/sort_order) and
// returns ordered [region, locations] groups for <optgroup>/section rendering.
export function groupByRegion(locations: FacultyLocation[]): Array<{ region: string; locations: FacultyLocation[] }> {
  const order: string[] = [];
  const map = new Map<string, FacultyLocation[]>();
  for (const loc of locations) {
    if (!map.has(loc.region)) {
      map.set(loc.region, []);
      order.push(loc.region);
    }
    map.get(loc.region)!.push(loc);
  }
  return order.map((region) => ({ region, locations: map.get(region)! }));
}

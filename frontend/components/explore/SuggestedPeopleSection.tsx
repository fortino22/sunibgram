import type { ExplorePerson } from '@/lib/services/explore';
import { PersonCard } from './PersonCard';

interface SuggestedPeopleSectionProps {
  people: ExplorePerson[];
}

export function SuggestedPeopleSection({ people }: SuggestedPeopleSectionProps) {
  if (people.length === 0) return null;

  return (
    <div>
      <div
        style={{
          padding: '16px 20px 8px',
          fontWeight: 700,
          fontSize: 20,
          color: 'var(--text)',
          letterSpacing: '-0.02em',
        }}
      >
        Who to follow
      </div>
      {people.map((person) => (
        <PersonCard key={person.userId} person={person} />
      ))}
    </div>
  );
}

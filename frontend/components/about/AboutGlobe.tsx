'use client';

import { Suspense, useMemo, useRef, useState } from 'react';
import { Canvas } from '@react-three/fiber';
import { Html, OrbitControls } from '@react-three/drei';
import { User } from 'lucide-react';
import * as THREE from 'three';

type Person = { name: string; lon: number; lat: number };

/* lon = longitude in degrees (0–360), lat = latitude in degrees
   (positive = upper hemisphere). Spread around the globe so ~half are
   visible on the camera-facing hemisphere at any rotation. */
const PEOPLE: Person[] = [
  { name: 'rina',   lon: 0,   lat: 58 },
  { name: 'aldo',   lon: 5,   lat: -38 },
  { name: 'tasya',  lon: 12,  lat: 18 },
  { name: 'nadine', lon: 22,  lat: -28 },
  { name: 'arif',   lon: 33,  lat: 40 },
  { name: 'adit',   lon: 45,  lat: -12 },
  { name: 'anya',   lon: 57,  lat: 22 },
  { name: 'nadya',  lon: 68,  lat: 50 },
  { name: 'fajar',  lon: 78,  lat: -35 },
  { name: 'raisa',  lon: 85,  lat: 8 },
  { name: 'putri',  lon: 90,  lat: 32 },
  { name: 'rendy',  lon: 100, lat: -22 },
  { name: 'dimas',  lon: 112, lat: 45 },
  { name: 'intan',  lon: 125, lat: 12 },
  { name: 'kevin',  lon: 135, lat: -42 },
  { name: 'lila',   lon: 145, lat: 28 },
  { name: 'sasha',  lon: 158, lat: -5 },
  { name: 'rian',   lon: 202, lat: 55 },
  { name: 'gilang', lon: 215, lat: -18 },
  { name: 'reza',   lon: 225, lat: 35 },
  { name: 'hadi',   lon: 235, lat: -32 },
  { name: 'ayla',   lon: 248, lat: 15 },
  { name: 'sari',   lon: 262, lat: -8 },
  { name: 'ayu',    lon: 270, lat: 48 },
  { name: 'lulu',   lon: 282, lat: 20 },
  { name: 'bayu',   lon: 292, lat: -15 },
  { name: 'bima',   lon: 305, lat: 38 },
  { name: 'daffa',  lon: 315, lat: 5 },
  { name: 'doni',   lon: 325, lat: -25 },
  { name: 'indra',  lon: 338, lat: 42 },
  { name: 'nisa',   lon: 350, lat: -40 },
];

const RADIUS = 1.5;
const DOT_COUNT = 900;

/* Fibonacci sphere — N points spread as evenly as possible on a sphere. */
function fibonacciSphere(count: number, r: number): Float32Array {
  const positions = new Float32Array(count * 3);
  const golden = Math.PI * (3 - Math.sqrt(5));
  for (let i = 0; i < count; i++) {
    const y = 1 - (i / (count - 1)) * 2;
    const ring = Math.sqrt(1 - y * y);
    const theta = golden * i;
    positions[i * 3]     = r * Math.cos(theta) * ring;
    positions[i * 3 + 1] = r * y;
    positions[i * 3 + 2] = r * Math.sin(theta) * ring;
  }
  return positions;
}

function sphereCoord(lonDeg: number, latDeg: number, r: number): [number, number, number] {
  const lon = (lonDeg * Math.PI) / 180;
  const lat = (latDeg * Math.PI) / 180;
  return [
    r * Math.cos(lat) * Math.sin(lon),
    r * Math.sin(lat),
    r * Math.cos(lat) * Math.cos(lon),
  ];
}

function Globe() {
  const occluderRef = useRef<THREE.Mesh>(null);

  const dotGeometry = useMemo(() => {
    const positions = fibonacciSphere(DOT_COUNT, RADIUS);
    const g = new THREE.BufferGeometry();
    g.setAttribute('position', new THREE.BufferAttribute(positions, 3));
    return g;
  }, []);

  return (
    /* Push the globe down inside the camera frustum so the bottom of
       the sphere falls past the canvas edge — combined with the
       section's `overflow: hidden`, this gives the bottom-clipped
       look from the reference. */
    <group position={[0, -0.85, 0]}>
      {/* Invisible occluder. drei's <Html occlude> raycasts against this
          mesh — labels swinging behind it fade out. Material writes
          neither colour nor depth, so the dots behind the back of the
          globe stay visible. */}
      <mesh ref={occluderRef}>
        <sphereGeometry args={[RADIUS * 0.95, 32, 32]} />
        <meshBasicMaterial colorWrite={false} depthWrite={false} />
      </mesh>

      {/* Dotted globe surface — a real 3D point cloud, distance-attenuated. */}
      <points geometry={dotGeometry}>
        <pointsMaterial
          color="#6b7280"
          size={0.03}
          sizeAttenuation
          transparent
          opacity={0.78}
          depthWrite={false}
        />
      </points>

      {/* HTML labels. drei projects each to screen space every frame and
          handles the billboard so the chip always faces the camera. */}
      {PEOPLE.map((p) => (
        <PersonLabel key={p.name} person={p} occluderRef={occluderRef} />
      ))}
    </group>
  );
}

/* One chip on the globe. Owns its own `occluded` boolean so React drives
   opacity instead of drei's default `display: none` toggle — that's what
   makes the fade-in/out work. */
function PersonLabel({
  person,
  occluderRef,
}: {
  person: Person;
  occluderRef: React.RefObject<THREE.Mesh | null>;
}) {
  const [occluded, setOccluded] = useState(false);

  return (
    <Html
      position={sphereCoord(person.lon, person.lat, RADIUS * 1.02)}
      center
      occlude={[occluderRef as React.RefObject<THREE.Object3D>]}
      onOcclude={setOccluded}
      zIndexRange={[100, 0]}
      style={{
        pointerEvents: 'none',
        opacity: occluded ? 0 : 1,
        transition: 'opacity 0.45s ease',
      }}
    >
      <span className="about-orbit-label-inner">
        <User size={14} />
        @{person.name}
      </span>
    </Html>
  );
}

export function AboutGlobe() {
  return (
    <Canvas
      dpr={[1, 2]}
      camera={{ position: [0, 0, 4.5], fov: 45 }}
      gl={{ antialias: true, alpha: true }}
      style={{ width: '100%', height: '100%', cursor: 'grab' }}
    >
      <Suspense fallback={null}>
        <Globe />
      </Suspense>
      {/* Drag to rotate in any direction; autoRotate keeps it spinning when idle. */}
      <OrbitControls
        enableZoom={false}
        enablePan={false}
        autoRotate
        autoRotateSpeed={0.6}
        enableDamping
        dampingFactor={0.08}
        rotateSpeed={0.8}
      />
    </Canvas>
  );
}

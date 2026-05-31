'use client';

import { useEffect, useRef, useState, useCallback, type RefObject } from 'react';
import { AnimatePresence, motion } from 'framer-motion';
import {
  Bookmark,
  Heart,
  MessageCircle,
  Repeat2,
  Send,
  TrendingUp,
  TrendingDown,
  Minus,
  Trophy,
  ArrowUp,
  ArrowDown,
  MoreHorizontal,
  ChevronDown,
  Globe,
  Image as ImageIcon,
  Film,
  Smile,
  BarChart3,
} from 'lucide-react';
import { Avatar } from '@/components/shared/Avatar';

/* ------------------------------------------------------------------ */
/*  Data                                                               */
/* ------------------------------------------------------------------ */

const COMPOSER_TEXT = "What's the best study spot at #Anggrek? 📚";

type Reply = { user: string; name: string; faculty: string; text: string };

const REPLIES: Reply[] = [
  { user: 'nadya', name: 'Nadya P.',  faculty: 'School of Design',     text: 'Anggrek library 1st floor — always quiet 🤫' },
  { user: 'maya',  name: 'Maya P.',   faculty: 'School of Info Sys.',  text: 'I love 1st floor too! Window seats >>> 🎧' },
];

const SUMMARY_BULLETS = [
  'A study spot question about Anggrek campus sparked a lively thread among Design and Comp. Sci. students.',
  'Students recommended the Anggrek library 1st floor and BMC window seats as top study spots.',
  'The thread gained 486 likes, 18 reposts, and 42 bookmarks — the campus clearly has opinions on this.',
];

type Trend = { name: string; posts: number };
const INITIAL_TRENDS: Trend[] = [
  { name: 'AnggrekVibes', posts: 1247 },
  { name: 'BinusBNPC',    posts: 894 },
  { name: 'StudyTips',    posts: 612 },
  { name: 'CampusFood',   posts: 478 },
  { name: 'FacultyChat',  posts: 312 },
  { name: 'ExamSeason',   posts: 215 },
];

type Leader = { user: string; name: string; faculty: string; score: number };
const INITIAL_LEADERS: Leader[] = [
  { user: 'rina',  name: 'Rina S.',   faculty: 'Design',     score: 92.3 },
  { user: 'kevin', name: 'Kevin H.',  faculty: 'Comp. Sci.', score: 88.7 },
  { user: 'maya',  name: 'Maya P.',   faculty: 'Info Sys.',  score: 84.1 },
  { user: 'reza',  name: 'Reza R.',   faculty: 'Math',       score: 79.5 },
  { user: 'putri', name: 'Putri W.',  faculty: 'Design',     score: 76.8 },
];

const SPEED = 1.6;
const T = (ms: number) => ms * SPEED;
const LOOP_MS = T(15_000);

/* ------------------------------------------------------------------ */
/*  Cursor position helper                                             */
/* ------------------------------------------------------------------ */

function getRelativeCenter(
  wrapperEl: HTMLDivElement | null,
  targetEl: HTMLElement | null,
): { x: number; y: number } {
  if (!wrapperEl || !targetEl) return { x: -40, y: -40 };
  const w = wrapperEl.getBoundingClientRect();
  const t = targetEl.getBoundingClientRect();
  return {
    x: t.left - w.left + t.width / 2,
    y: t.top - w.top + t.height / 2,
  };
}

/* ------------------------------------------------------------------ */
/*  Main section                                                       */
/* ------------------------------------------------------------------ */

export function AboutThreadFlow() {
  /* Existing state */
  const [typedChars, setTypedChars] = useState(0);
  const [posted, setPosted] = useState(false);
  const [likes, setLikes] = useState(0);
  const [reposts, setReposts] = useState(0);
  const [comments, setComments] = useState(0);
  const [bookmarks, setBookmarks] = useState(0);
  const [liked, setLiked] = useState(false);
  const [replyLiked, setReplyLiked] = useState(false);
  const [bookmarked, setBookmarked] = useState(false);
  const [replies, setReplies] = useState<Reply[]>([]);

  const [trends, setTrends] = useState<Trend[]>(INITIAL_TRENDS);
  const [leaders, setLeaders] = useState<Leader[]>(INITIAL_LEADERS);

  /* New state for cursor animation & summary */
  const [cursorPos, setCursorPos] = useState({ x: -40, y: -40 });
  const [cursorVisible, setCursorVisible] = useState(false);
  const [cursorClicking, setCursorClicking] = useState(false);
  const [composerFocused, setComposerFocused] = useState(false);
  const [foxGlowing, setFoxGlowing] = useState(false);
  const [showSummary, setShowSummary] = useState(false);
  const [visibleBullets, setVisibleBullets] = useState(0);

  /* Refs for cursor targeting */
  const wrapperRef = useRef<HTMLDivElement>(null);
  const composerRef = useRef<HTMLDivElement>(null);
  const postBtnRef = useRef<HTMLButtonElement>(null);
  const likeRef = useRef<HTMLSpanElement>(null);
  const replyLikeRef = useRef<HTMLSpanElement>(null);
  const bookmarkRef = useRef<HTMLSpanElement>(null);
  const foxRef = useRef<HTMLButtonElement>(null);

  const moveCursorTo = useCallback(
    (ref: RefObject<HTMLElement | null>) => {
      const pos = getRelativeCenter(wrapperRef.current, ref.current);
      setCursorPos(pos);
    },
    [],
  );

  /* ----- Thread cycle ---------------------------------------------- */
  useEffect(() => {
    let mounted = true;
    let typeInterval: ReturnType<typeof setInterval> | null = null;
    const timeouts: ReturnType<typeof setTimeout>[] = [];

    function click() {
      setCursorClicking(true);
      setTimeout(() => { if (mounted) setCursorClicking(false); }, 200);
    }

    function runCycle() {
      if (!mounted) return;

      timeouts.forEach((t) => clearTimeout(t));
      timeouts.length = 0;
      if (typeInterval) {
        clearInterval(typeInterval);
        typeInterval = null;
      }

      /* Reset everything */
      setTypedChars(0);
      setPosted(false);
      setLikes(0);
      setReposts(0);
      setComments(0);
      setBookmarks(0);
      setLiked(false);
      setReplyLiked(false);
      setBookmarked(false);
      setReplies([]);
      setCursorVisible(false);
      setCursorClicking(false);
      setComposerFocused(false);
      setFoxGlowing(false);
      setShowSummary(false);
      setVisibleBullets(0);

      /* Cursor appears and moves to composer */
      timeouts.push(setTimeout(() => {
        setCursorVisible(true);
        moveCursorTo(composerRef);
      }, T(200)));

      /* Cursor clicks composer text area → focus + typing begins */
      timeouts.push(setTimeout(() => {
        click();
        setComposerFocused(true);
      }, T(500)));

      /* Start typing */
      timeouts.push(setTimeout(() => {
        typeInterval = setInterval(() => {
          setTypedChars((p) => {
            if (p >= COMPOSER_TEXT.length) {
              if (typeInterval) {
                clearInterval(typeInterval);
                typeInterval = null;
              }
              return p;
            }
            return p + 1;
          });
        }, T(55));
      }, T(700)));

      /* Typing done */
      timeouts.push(setTimeout(() => {
        setTypedChars(COMPOSER_TEXT.length);
      }, T(2600)));

      /* Cursor moves to Post button */
      timeouts.push(setTimeout(() => {
        moveCursorTo(postBtnRef);
      }, T(2800)));

      /* Cursor clicks Post → composer transforms into posted thread */
      timeouts.push(setTimeout(() => {
        click();
        setPosted(true);
      }, T(3000)));

      /* Cursor moves to like button on posted card */
      timeouts.push(setTimeout(() => {
        moveCursorTo(likeRef);
      }, T(3400)));

      /* Cursor clicks like */
      timeouts.push(setTimeout(() => {
        click();
        setLikes(7);
        setLiked(true);
      }, T(3600)));

      /* Likes grow */
      timeouts.push(setTimeout(() => { setLikes(23); setReposts(1); }, T(4200)));

      /* First reply + comment */
      timeouts.push(setTimeout(() => {
        setLikes(47);
        setComments(1);
        setReplies((r) => [...r, REPLIES[0]]);
      }, T(4900)));

      /* Cursor moves to reply's like button */
      timeouts.push(setTimeout(() => {
        moveCursorTo(replyLikeRef);
      }, T(5300)));

      /* Cursor clicks reply like */
      timeouts.push(setTimeout(() => {
        click();
        setLikes(89);
        setReposts(3);
        setReplyLiked(true);
      }, T(5500)));

      /* More growth + second reply */
      timeouts.push(setTimeout(() => {
        setLikes(178);
        setReposts(6);
        setComments(2);
        setReplies((r) => [...r, REPLIES[1]]);
      }, T(6700)));

      /* Cursor moves to bookmark */
      timeouts.push(setTimeout(() => {
        moveCursorTo(bookmarkRef);
      }, T(7500)));

      /* Cursor clicks bookmark */
      timeouts.push(setTimeout(() => {
        click();
        setLikes(312);
        setReposts(11);
        setBookmarks(42);
        setBookmarked(true);
      }, T(7700)));

      /* Likes + reposts keep growing */
      timeouts.push(setTimeout(() => {
        setLikes(486);
        setReposts(18);
      }, T(9000)));

      /* Cursor moves to fox button */
      timeouts.push(setTimeout(() => {
        moveCursorTo(foxRef);
      }, T(9500)));

      /* Cursor clicks fox button → glow + summary */
      timeouts.push(setTimeout(() => {
        click();
        setFoxGlowing(true);
      }, T(9700)));

      /* Summary modal appears */
      timeouts.push(setTimeout(() => {
        setShowSummary(true);
        setCursorVisible(false);
      }, T(10000)));

      /* Stream summary bullets */
      timeouts.push(setTimeout(() => setVisibleBullets(1), T(10300)));
      timeouts.push(setTimeout(() => setVisibleBullets(2), T(10900)));
      timeouts.push(setTimeout(() => setVisibleBullets(3), T(11500)));

      /* Summary modal closes */
      timeouts.push(setTimeout(() => {
        setShowSummary(false);
        setFoxGlowing(false);
      }, T(13300)));

      /* Next cycle */
      timeouts.push(setTimeout(runCycle, LOOP_MS));
    }
    runCycle();

    return () => {
      mounted = false;
      if (typeInterval) clearInterval(typeInterval);
      timeouts.forEach((t) => clearTimeout(t));
    };
  }, [moveCursorTo]);

  /* ----- Trending shuffle (every 2.6 s the list re-sorts) ----------- */
  useEffect(() => {
    const id = setInterval(() => {
      setTrends((current) => {
        const next = current.map((t) => ({
          ...t,
          posts: t.posts + Math.floor(Math.random() * 80 + 5),
        }));
        next.sort((a, b) => b.posts - a.posts);
        return next;
      });
    }, 2600);
    return () => clearInterval(id);
  }, []);

  /* ----- Leaderboard shuffle (every 2.4 s scores tick) -------------- */
  useEffect(() => {
    const id = setInterval(() => {
      setLeaders((current) => {
        const next = current.map((l) => ({
          ...l,
          score: Math.max(0, +(l.score + (Math.random() * 4 - 1.4)).toFixed(1)),
        }));
        next.sort((a, b) => b.score - a.score);
        return next;
      });
    }, 2400);
    return () => clearInterval(id);
  }, []);

  return (
    <section className="about-container about-thread-section">
      <div className="thread-flow-grid">
        <TrendingPanel trends={trends} />

        <ThreadStage
          posted={posted}
          composerFocused={composerFocused}
          displayText={COMPOSER_TEXT.slice(0, typedChars)}
          typingDone={typedChars >= COMPOSER_TEXT.length}
          typedChars={typedChars}
          likes={likes}
          reposts={reposts}
          comments={comments}
          bookmarks={bookmarks}
          liked={liked}
          replyLiked={replyLiked}
          bookmarked={bookmarked}
          replies={replies}
          wrapperRef={wrapperRef}
          composerRef={composerRef}
          postBtnRef={postBtnRef}
          likeRef={likeRef}
          replyLikeRef={replyLikeRef}
          bookmarkRef={bookmarkRef}
          foxRef={foxRef}
          cursorPos={cursorPos}
          cursorVisible={cursorVisible}
          cursorClicking={cursorClicking}
          foxGlowing={foxGlowing}
          showSummary={showSummary}
          visibleBullets={visibleBullets}
        />

        <LeaderboardPanel leaders={leaders} />
      </div>
    </section>
  );
}

/* ------------------------------------------------------------------ */
/*  Left — Trending hashtags (animated re-sort)                        */
/* ------------------------------------------------------------------ */

function TrendingPanel({ trends }: { trends: Trend[] }) {
  return (
    <aside className="flow-card flow-trending">
      <header className="flow-card-header">
        <h3 className="flow-card-title">Trending on campus</h3>
      </header>
      <motion.ul layout className="flow-trending-list">
        {trends.map((t, i) => (
          <TrendingRow key={t.name} trend={t} rank={i + 1} />
        ))}
      </motion.ul>
    </aside>
  );
}

function TrendingRow({ trend, rank }: { trend: Trend; rank: number }) {
  /* Track previous rank to know if it moved up or down. */
  const prevRankRef = useRef<number>(rank);
  const [movement, setMovement] = useState<'up' | 'down' | 'same'>('same');

  useEffect(() => {
    if (prevRankRef.current === rank) return;
    setMovement(rank < prevRankRef.current ? 'up' : 'down');
    prevRankRef.current = rank;
    const id = setTimeout(() => setMovement('same'), 1200);
    return () => clearTimeout(id);
  }, [rank]);

  const Icon = movement === 'up' ? TrendingUp : movement === 'down' ? TrendingDown : Minus;
  const iconColor =
    movement === 'up' ? 'var(--green)' :
    movement === 'down' ? 'var(--coral)' :
    'var(--text-tertiary)';

  return (
    <motion.li
      layout
      transition={{ type: 'spring', damping: 26, stiffness: 220 }}
      className="flow-trending-row"
    >
      <div className="flow-trending-info">
        <p className="flow-trending-meta">Trending · Campus</p>
        <p className="flow-trending-tag">#{trend.name}</p>
        <p className="flow-trending-posts">{trend.posts.toLocaleString()} posts</p>
      </div>
      <motion.span
        key={`${trend.name}-${movement}`}
        className="flow-trending-delta"
        initial={{ opacity: 0, scale: 0.6 }}
        animate={{ opacity: 1, scale: 1 }}
        transition={{ duration: 0.25 }}
        style={{ color: iconColor }}
      >
        <Icon size={14} strokeWidth={2.5} />
      </motion.span>
    </motion.li>
  );
}

/* ------------------------------------------------------------------ */
/*  Centre — Thread card (matches ProfileThreadCard structure)         */
/* ------------------------------------------------------------------ */

function ThreadStage({
  posted,
  composerFocused,
  displayText,
  typingDone,
  typedChars,
  likes,
  reposts,
  comments,
  bookmarks,
  liked,
  replyLiked,
  bookmarked,
  replies,
  wrapperRef,
  composerRef,
  postBtnRef,
  likeRef,
  replyLikeRef,
  bookmarkRef,
  foxRef,
  cursorPos,
  cursorVisible,
  cursorClicking,
  foxGlowing,
  showSummary,
  visibleBullets,
}: {
  posted: boolean;
  composerFocused: boolean;
  displayText: string;
  typingDone: boolean;
  typedChars: number;
  likes: number;
  reposts: number;
  comments: number;
  bookmarks: number;
  liked: boolean;
  replyLiked: boolean;
  bookmarked: boolean;
  replies: Reply[];
  wrapperRef: RefObject<HTMLDivElement | null>;
  composerRef: RefObject<HTMLDivElement | null>;
  postBtnRef: RefObject<HTMLButtonElement | null>;
  likeRef: RefObject<HTMLSpanElement | null>;
  replyLikeRef: RefObject<HTMLSpanElement | null>;
  bookmarkRef: RefObject<HTMLSpanElement | null>;
  foxRef: RefObject<HTMLButtonElement | null>;
  cursorPos: { x: number; y: number };
  cursorVisible: boolean;
  cursorClicking: boolean;
  foxGlowing: boolean;
  showSummary: boolean;
  visibleBullets: number;
}) {
  return (
    <div className="flow-thread-wrapper" ref={wrapperRef} style={{ position: 'relative' }}>
      <AnimatePresence mode="wait" initial={false}>
        {!posted ? (
          <motion.div
            key="composer"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0, y: -8 }}
            transition={{ duration: 0.25, ease: 'easeOut' }}
          >
            <ThreadComposer
              composerRef={composerRef}
              postBtnRef={postBtnRef}
              displayText={displayText}
              typingDone={typingDone}
              typedChars={typedChars}
              focused={composerFocused}
            />
          </motion.div>
        ) : (
          <motion.article
            key="posted"
            initial={{ opacity: 0, y: 8 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.3, ease: 'easeOut' }}
            className="flow-thread-card"
          >
            <div className="flow-thread-row">
              <Avatar username="you" avatarUrl={null} size="md" />
              <div className="flow-thread-main">
                <div className="flow-thread-headline">
                  <span className="flow-thread-name">You</span>
                  <span className="flow-thread-handle">@you</span>
                  <span className="flow-thread-dot">·</span>
                  <span className="flow-thread-time">just now</span>

                  {/* Fox button + three-dots */}
                  <div className="flow-thread-actions">
                    <button
                      ref={foxRef}
                      className={`flow-thread-fox${foxGlowing ? ' flow-thread-fox-glow' : ''}`}
                      tabIndex={-1}
                      aria-label="AI summary"
                    >
                      {/* eslint-disable-next-line @next/next/no-img-element */}
                      <img src="/assets/fox.png" alt="Fox" />
                    </button>
                    <span className="flow-thread-dots">
                      <MoreHorizontal size={18} strokeWidth={2} />
                    </span>
                  </div>
                </div>
                <div className="flow-thread-faculty">School of Design</div>

                <div className="flow-thread-content">
                  <ThreadText text={displayText} showCursor={false} />
                </div>

                <ActionBarMock
                  likes={likes}
                  reposts={reposts}
                  comments={comments}
                  bookmarks={bookmarks}
                  liked={liked}
                  reposted={false}
                  bookmarked={bookmarked}
                  active={posted}
                  likeRef={likeRef}
                  repostRef={null}
                  bookmarkRef={bookmarkRef}
                />
              </div>
            </div>
          </motion.article>
        )}
      </AnimatePresence>

      {/* Reply composer (appears once the thread is posted) */}
      {posted && (
        <div className="flow-reply-composer">
          <Avatar username="you" avatarUrl={null} size="sm" />
          <span className="flow-reply-placeholder">Post your reply to @you</span>
          <span className="flow-reply-btn">Reply</span>
        </div>
      )}

      {/* Reply previews — slide in like real ProfileThreadCard replies. */}
      <AnimatePresence initial={false}>
        {replies.map((r) => (
          <motion.article
            key={r.user}
            layout
            initial={{ opacity: 0, y: 14 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.35, ease: 'easeOut' }}
            className="flow-thread-card flow-thread-reply"
          >
            <div className="flow-thread-row">
              <Avatar username={r.user} avatarUrl={null} size="md" />
              <div className="flow-thread-main">
                <div className="flow-thread-headline">
                  <span className="flow-thread-name">{r.name}</span>
                  <span className="flow-thread-handle">@{r.user}</span>
                  <span className="flow-thread-dot">·</span>
                  <span className="flow-thread-time">now</span>

                  {/* Fox + dots on replies too */}
                  <div className="flow-thread-actions">
                    <span className="flow-thread-fox">
                      {/* eslint-disable-next-line @next/next/no-img-element */}
                      <img src="/assets/fox.png" alt="Fox" />
                    </span>
                    <span className="flow-thread-dots">
                      <MoreHorizontal size={18} strokeWidth={2} />
                    </span>
                  </div>
                </div>
                <div className="flow-thread-faculty">{r.faculty}</div>
                <div className="flow-thread-content">
                  <ThreadText text={r.text} showCursor={false} />
                </div>

                {/* Static action bar on replies */}
                <ReplyActionBar 
                  isLiked={r.user === 'nadya' && replyLiked}
                  likeRef={r.user === 'nadya' ? replyLikeRef : null} 
                />
              </div>
            </div>
          </motion.article>
        ))}
      </AnimatePresence>

      {/* Animated cursor overlay */}
      <AnimatedCursor
        pos={cursorPos}
        visible={cursorVisible}
        clicking={cursorClicking}
      />

      {/* AI Summary modal — appears directly below the fox icon */}
      <SummaryModal visible={showSummary} visibleBullets={visibleBullets} />
    </div>
  );
}

/* ------------------------------------------------------------------ */
/*  Animated cursor                                                    */
/* ------------------------------------------------------------------ */

function AnimatedCursor({
  pos,
  visible,
  clicking,
}: {
  pos: { x: number; y: number };
  visible: boolean;
  clicking: boolean;
}) {
  return (
    <motion.div
      className="flow-animated-cursor"
      animate={{
        x: pos.x,
        y: pos.y,
        scale: clicking ? 0.8 : 1,
        opacity: visible ? 1 : 0,
      }}
      transition={{
        x: { type: 'tween', duration: 0.5, ease: 'easeInOut' },
        y: { type: 'tween', duration: 0.5, ease: 'easeInOut' },
        scale: { duration: 0.1 },
        opacity: { duration: 0.2 },
      }}
    >
      {/* Classic pointer cursor SVG */}
      <svg width="18" height="22" viewBox="0 0 18 22" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path
          d="M1.5 1L1.5 17L5.75 13L9.5 20.5L12 19.25L8.25 11.75L13.5 11.75L1.5 1Z"
          fill="var(--text)"
          stroke="var(--bg)"
          strokeWidth="1.5"
          strokeLinejoin="round"
        />
      </svg>

      {/* Click ring effect */}
      <AnimatePresence>
        {clicking && (
          <motion.div
            key="ring"
            className="flow-cursor-click-ring"
            initial={{ scale: 0.5, opacity: 1 }}
            animate={{ scale: 2.5, opacity: 0 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.4, ease: 'easeOut' }}
          />
        )}
      </AnimatePresence>
    </motion.div>
  );
}

/* ------------------------------------------------------------------ */
/*  AI Summary modal                                                   */
/* ------------------------------------------------------------------ */

function SummaryModal({
  visible,
  visibleBullets,
}: {
  visible: boolean;
  visibleBullets: number;
}) {
  return (
    <AnimatePresence>
      {visible && (
        <motion.div
          className="flow-summary-modal"
          initial={{ opacity: 0, y: 20, scale: 0.95 }}
          animate={{ opacity: 1, y: 0, scale: 1 }}
          exit={{ opacity: 0, y: 20, scale: 0.95 }}
          transition={{ duration: 0.3, ease: 'easeOut' }}
        >
          {/* Header */}
          <div className="flow-summary-header">
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img src="/assets/fox.png" alt="Fox" className="flow-summary-fox-icon" />
            <span className="flow-summary-title">Summary</span>
            <ChevronDown size={16} className="flow-summary-chevron" />
          </div>

          {/* Full thread section */}
          <div className="flow-summary-section-label">Full Thread</div>
          <div className="flow-summary-thread-preview">
            <Avatar username="you" avatarUrl={null} size="sm" />
            <div className="flow-summary-preview-meta">
              <div>
                <span className="flow-summary-preview-name">You</span>
                <span className="flow-summary-preview-handle">@you · just now</span>
              </div>
              <p className="flow-summary-preview-text">{COMPOSER_TEXT}</p>
            </div>
          </div>

          {/* Summary section */}
          <div className="flow-summary-section-label">Summary</div>
          <ul className="flow-summary-bullets">
            {SUMMARY_BULLETS.slice(0, visibleBullets).map((bullet, i) => (
              <motion.li
                key={i}
                className="flow-summary-bullet"
                initial={{ opacity: 0, y: 8 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.3 }}
              >
                {bullet}
                {i === visibleBullets - 1 && (
                  <span className="flow-summary-streaming-cursor"> ▍</span>
                )}
              </motion.li>
            ))}
          </ul>
        </motion.div>
      )}
    </AnimatePresence>
  );
}

/* ------------------------------------------------------------------ */
/*  Thread composer (initial state — pre-post)                          */
/* ------------------------------------------------------------------ */

function ThreadComposer({
  composerRef,
  postBtnRef,
  displayText,
  typingDone,
  typedChars,
  focused,
}: {
  composerRef: RefObject<HTMLDivElement | null>;
  postBtnRef: RefObject<HTMLButtonElement | null>;
  displayText: string;
  typingDone: boolean;
  typedChars: number;
  focused: boolean;
}) {
  const hasText = typedChars > 0;
  return (
    <div className="flow-composer">
      <div className="flow-composer-top">
        <Avatar username="you" avatarUrl={null} size="md" />
        <div className="flow-composer-main">
          <button type="button" tabIndex={-1} className="flow-composer-audience">
            <Globe size={13} strokeWidth={2.4} />
            <span>Everyone</span>
            <ChevronDown size={12} strokeWidth={2.4} />
          </button>
          <div ref={composerRef} className="flow-composer-textarea">
            {hasText ? (
              <ThreadText text={displayText} showCursor={!typingDone} />
            ) : (
              <>
                {focused && <span className="flow-thread-cursor">|</span>}
                <span className="flow-composer-placeholder">What is happening?!</span>
              </>
            )}
          </div>
          <div className="flow-composer-reply-info">
            <Globe size={12} strokeWidth={2.4} />
            <span>Everyone can reply</span>
          </div>
        </div>
      </div>
      <div className="flow-composer-divider" />
      <div className="flow-composer-bottom">
        <div className="flow-composer-actions">
          <span className="flow-composer-action"><ImageIcon size={18} strokeWidth={2} /></span>
          <span className="flow-composer-action"><Film size={18} strokeWidth={2} /></span>
          <span className="flow-composer-action"><Smile size={18} strokeWidth={2} /></span>
          <span className="flow-composer-action"><BarChart3 size={18} strokeWidth={2} /></span>
        </div>
        <button
          ref={postBtnRef}
          type="button"
          tabIndex={-1}
          className={`flow-composer-post${hasText ? ' flow-composer-post-ready' : ''}`}
        >
          Post
        </button>
      </div>
    </div>
  );
}

/* ------------------------------------------------------------------ */
/*  Thread text with hashtag highlights & blinking cursor               */
/* ------------------------------------------------------------------ */

/* Renders text with #hashtag highlights and a blinking cursor. */
function ThreadText({ text, showCursor }: { text: string; showCursor: boolean }) {
  const parts = text.split(/(\s+)/);
  return (
    <span className="flow-thread-text">
      {parts.map((p, i) =>
        p.startsWith('#') ? (
          <span key={i} className="flow-thread-hashtag">{p}</span>
        ) : (
          <span key={i}>{p}</span>
        )
      )}
      {showCursor && <span className="flow-thread-cursor">|</span>}
    </span>
  );
}

/* ------------------------------------------------------------------ */
/*  Action bar mock with refs for cursor targeting                     */
/* ------------------------------------------------------------------ */

/* Inline action bar mimicking ActionBar.tsx visually. */
function ActionBarMock({
  likes,
  reposts,
  comments,
  bookmarks,
  liked,
  reposted,
  bookmarked,
  active,
  likeRef,
  repostRef,
  bookmarkRef,
}: {
  likes: number;
  reposts: number;
  comments: number;
  bookmarks: number;
  liked: boolean;
  reposted: boolean;
  bookmarked: boolean;
  active: boolean;
  likeRef: RefObject<HTMLSpanElement | null>;
  repostRef: RefObject<HTMLSpanElement | null> | null;
  bookmarkRef: RefObject<HTMLSpanElement | null>;
}) {
  return (
    <div className="flow-action-bar" data-active={active ? 'true' : 'false'}>
      <span className="flow-action" style={{ color: 'var(--text-tertiary)' }}>
        <MessageCircle size={18} strokeWidth={2} />
        <span className="flow-action-count">{comments}</span>
      </span>
      <span
        ref={repostRef}
        className="flow-action"
        style={{ color: reposted ? 'var(--green)' : 'var(--text-tertiary)' }}
      >
        <Repeat2 size={18} strokeWidth={2} />
        <span className="flow-action-count">{reposts}</span>
      </span>
      <motion.span
        ref={likeRef}
        className="flow-action"
        animate={liked ? { scale: [1, 1.35, 1] } : { scale: 1 }}
        transition={{ duration: 0.35 }}
        style={{ color: liked ? 'var(--coral)' : 'var(--text-tertiary)' }}
      >
        <Heart
          size={18}
          strokeWidth={2}
          fill={liked ? 'var(--coral)' : 'none'}
        />
        <span className="flow-action-count">{likes}</span>
      </motion.span>
      <span
        ref={bookmarkRef}
        className="flow-action"
        style={{ color: bookmarked ? 'var(--accent)' : 'var(--text-tertiary)' }}
      >
        <Bookmark
          size={18}
          strokeWidth={2}
          fill={bookmarked ? 'var(--accent)' : 'none'}
        />
        {bookmarks > 0 && <span className="flow-action-count">{bookmarks}</span>}
      </span>
      <span className="flow-action" style={{ color: 'var(--text-tertiary)' }}>
        <Send size={18} strokeWidth={2} />
      </span>
    </div>
  );
}

/* Static action bar for reply cards — same icons, all zeroed out. */
function ReplyActionBar({ 
  isLiked = false, 
  likeRef 
}: { 
  isLiked?: boolean; 
  likeRef?: RefObject<HTMLSpanElement | null> 
}) {
  return (
    <div className="flow-action-bar" data-active="true">
      <span className="flow-action" style={{ color: 'var(--text-tertiary)' }}>
        <MessageCircle size={18} strokeWidth={2} />
        <span className="flow-action-count">0</span>
      </span>
      <span className="flow-action" style={{ color: 'var(--text-tertiary)' }}>
        <Repeat2 size={18} strokeWidth={2} />
        <span className="flow-action-count">0</span>
      </span>
      <motion.span 
        ref={likeRef}
        className="flow-action" 
        animate={isLiked ? { scale: [1, 1.35, 1] } : { scale: 1 }}
        transition={{ duration: 0.35 }}
        style={{ color: isLiked ? 'var(--coral)' : 'var(--text-tertiary)' }}
      >
        <Heart 
          size={18} 
          strokeWidth={2} 
          fill={isLiked ? 'var(--coral)' : 'none'} 
        />
        <span className="flow-action-count">{isLiked ? 1 : 0}</span>
      </motion.span>
      <span className="flow-action" style={{ color: 'var(--text-tertiary)' }}>
        <Bookmark size={18} strokeWidth={2} />
      </span>
      <span className="flow-action" style={{ color: 'var(--text-tertiary)' }}>
        <Send size={18} strokeWidth={2} />
      </span>
    </div>
  );
}

/* ------------------------------------------------------------------ */
/*  Right — Leaderboard (animated re-sort with score ticks)            */
/* ------------------------------------------------------------------ */

function LeaderboardPanel({ leaders }: { leaders: Leader[] }) {
  return (
    <aside className="flow-card flow-leaderboard">
      <header className="flow-card-header">
        <h3 className="flow-card-title">
          <Trophy size={16} strokeWidth={2.2} style={{ color: 'var(--amber)' }} />
          Campus leaderboard
        </h3>
      </header>
      <motion.ul layout className="flow-leaderboard-list">
        {leaders.map((l, i) => (
          <LeaderRow key={l.user} leader={l} rank={i + 1} />
        ))}
      </motion.ul>
    </aside>
  );
}

function LeaderRow({ leader, rank }: { leader: Leader; rank: number }) {
  const prevScoreRef = useRef<number>(leader.score);
  const prevRankRef = useRef<number>(rank);
  const [scoreDelta, setScoreDelta] = useState<'up' | 'down' | 'same'>('same');

  useEffect(() => {
    if (Math.abs(leader.score - prevScoreRef.current) < 0.05) return;
    setScoreDelta(leader.score > prevScoreRef.current ? 'up' : 'down');
    prevScoreRef.current = leader.score;
    const id = setTimeout(() => setScoreDelta('same'), 900);
    return () => clearTimeout(id);
  }, [leader.score]);

  useEffect(() => {
    prevRankRef.current = rank;
  }, [rank]);

  return (
    <motion.li
      layout
      transition={{ type: 'spring', damping: 26, stiffness: 220 }}
      className="flow-leader-row"
      data-rank={rank}
    >
      <span className="flow-leader-rank">{rank}</span>
      <div className="flow-leader-user">
        <Avatar username={leader.user} avatarUrl={null} size={32} />
        <div className="flow-leader-meta">
          <span className="flow-leader-name">{leader.name}</span>
          <span className="flow-leader-faculty">@{leader.user} · {leader.faculty}</span>
        </div>
      </div>
      <div className="flow-leader-score">
        <span className="flow-leader-score-value">{leader.score.toFixed(1)}</span>
        <AnimatePresence mode="wait">
          {scoreDelta !== 'same' && (
            <motion.span
              key={scoreDelta + leader.score}
              initial={{ opacity: 0, y: scoreDelta === 'up' ? 6 : -6 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0 }}
              transition={{ duration: 0.25 }}
              className="flow-leader-arrow"
              style={{ color: scoreDelta === 'up' ? 'var(--green)' : 'var(--coral)' }}
            >
              {scoreDelta === 'up' ? <ArrowUp size={13} /> : <ArrowDown size={13} />}
            </motion.span>
          )}
        </AnimatePresence>
      </div>
    </motion.li>
  );
}

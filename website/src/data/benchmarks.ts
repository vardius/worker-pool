// Results behind the benchmark charts and tables, updated together. The suite reruns with the commands on the
// Benchmarks page; the comparison comes from a harness that runs the old and new versions side by side.

export const environment = 'Apple M4 Max, Go 1.27.0, median of 5 runs';

/** The two versions the charts compare. */
export const versions = { before: 'v2', after: 'v3' } as const;

export type Group = 'delegate' | 'job';

export interface Result {
	ns: number;
	allocs: number;
}

export interface Comparison {
	scenario: string;
	before: Result;
	after: Result;
}

/**
 * v2 against v3 on the same harness, one int job per op, with workers that do nothing.
 * Delegate runs into a queue with room for every job; a job runs through a queue of 1024, at the workers' pace.
 */
export const comparison: Record<Group, Comparison[]> = {
	delegate: [
		{ scenario: 'queue has room', before: { ns: 32.6, allocs: 2 }, after: { ns: 11.4, allocs: 0 } },
		{ scenario: 'queue has room, 16 callers at once', before: { ns: 86.0, allocs: 2 }, after: { ns: 35.4, allocs: 0 } },
	],
	job: [
		{ scenario: '1 worker', before: { ns: 215.9, allocs: 2 }, after: { ns: 38.8, allocs: 0 } },
		{ scenario: '16 workers', before: { ns: 412.1, allocs: 2 }, after: { ns: 110.1, allocs: 0 } },
		{ scenario: '1 worker, 16 callers at once', before: { ns: 372.3, allocs: 2 }, after: { ns: 86.9, allocs: 0 } },
		{ scenario: '16 workers, 16 callers at once', before: { ns: 258.3, allocs: 2 }, after: { ns: 55.4, allocs: 0 } },
	],
};

export interface SuiteResult {
	benchmark: string;
	what: string;
	ns: number;
	bytes: number;
	allocs: number;
}

/** v3's own benchmark suite, `go test -bench .`. */
export const suite: SuiteResult[] = [
	{ benchmark: 'Delegate', what: 'Into a queue with room, no worker', ns: 11.0, bytes: 0, allocs: 0 },
	{ benchmark: 'Job/workers=1', what: 'One caller, one worker, a queue of 1024', ns: 38.0, bytes: 0, allocs: 0 },
	{ benchmark: 'Job/workers=16', what: 'One caller, sixteen workers', ns: 104.1, bytes: 0, allocs: 0 },
	{ benchmark: 'JobParallel', what: 'Sixteen callers, sixteen workers', ns: 46.2, bytes: 0, allocs: 0 },
];

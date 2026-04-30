export function firstSentence(input: string | undefined, max = 140): string {
  if (!input) return '';
  const trimmed = input.replace(/\s+/g, ' ').trim();
  if (!trimmed) return '';
  const sentenceEnd = trimmed.search(/(?<=[.?!])\s/);
  const candidate =
    sentenceEnd > 0 && sentenceEnd <= max ? trimmed.slice(0, sentenceEnd + 1) : trimmed;
  if (candidate.length <= max) return candidate;
  return candidate.slice(0, max - 1).trimEnd() + '…';
}

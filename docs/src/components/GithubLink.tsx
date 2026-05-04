import { IconButton } from '@entur/button';
import { Tooltip } from '@entur/tooltip';
import { GithubIcon } from '@entur/icons';

export function GithubLink() {
  return (
    <Tooltip content="View on GitHub" placement="bottom">
      <IconButton
        as="a"
        href="https://github.com/entur/ai"
        target="_blank"
        rel="noopener noreferrer"
        aria-label="View entur/ai on GitHub"
      >
        <GithubIcon aria-hidden />
      </IconButton>
    </Tooltip>
  );
}

export default GithubLink;

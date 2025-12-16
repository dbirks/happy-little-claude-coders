FROM debian:bookworm-slim

# Install base dependencies
RUN apt-get update && apt-get install -y \
    curl \
    git \
    ca-certificates \
    gnupg \
    && rm -rf /var/lib/apt/lists/*

# Install Node.js 24
RUN curl -fsSL https://deb.nodesource.com/setup_24.x | bash - \
    && apt-get install -y nodejs \
    && rm -rf /var/lib/apt/lists/*

# Install GitHub CLI
RUN (curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | gpg --dearmor -o /usr/share/keyrings/githubcli-archive-keyring.gpg) \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" > /etc/apt/sources.list.d/github-cli.list \
    && apt-get update \
    && apt-get install -y gh \
    && rm -rf /var/lib/apt/lists/*

# Install Claude Code CLI
RUN curl -fsSL https://claude.ai/install.sh | bash

# Install pnpm
RUN npm install -g pnpm@9.15.2

# Install Happy CLI
RUN pnpm install -g happy-coder

# Create workspace directory
RUN mkdir -p /workspace /scripts

# Copy entrypoint and helper scripts
COPY scripts/entrypoint.sh /scripts/entrypoint.sh
COPY scripts/clone-repos.sh /usr/local/bin/clone-repos

RUN chmod +x /scripts/entrypoint.sh /usr/local/bin/clone-repos

# Set working directory
WORKDIR /workspace

# Default entrypoint
ENTRYPOINT ["/scripts/entrypoint.sh"]
CMD ["bash"]

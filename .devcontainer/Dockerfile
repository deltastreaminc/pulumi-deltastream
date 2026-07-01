FROM mcr.microsoft.com/devcontainers/base:ubuntu

# Install mise
RUN curl https://mise.run | sh
ENV PATH="/root/.local/bin:/root/.local/share/mise/shims:${PATH}"

# Activate mise in shell for interactive sessions and add shims to PATH
RUN echo 'eval "$(mise activate bash)"' >> ~/.bashrc && \
    echo 'export PATH="/root/.local/share/mise/shims:$PATH"' >> ~/.profile

# Set mise to trust all config files in the workspace automatically.
# MISE_YES=1 suppresses interactive prompts (equivalent to --yes on all commands).
ENV MISE_YES=1

#!/usr/bin/with-contenv bashio
# ==============================================================================
# Handle corefile generation
# ==============================================================================

# Generate corefile
if ! tempio \
        -conf /config/coredns.json \
        -template /usr/share/tempio/corefile \
        -out /etc/corefile
then
    bashio::log.error "Corefile fails to generate. Use fallback corefile!"
fi

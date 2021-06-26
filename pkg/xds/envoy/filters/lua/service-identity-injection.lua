function envoy_on_request(request_handle)
    if request_handle:connection():ssl() and request_handle:streamInfo():downstreamSslConnection() then
        local cert_uris = request_handle:streamInfo():downstreamSslConnection():uriSanPeerCertificate()
        for _, uri in pairs(cert_uris) do
            if uri:find("^spiffe://") ~= nil then
                request_handle:headers():add("X-Kuma-Forwarded-Client-Cert", uri)
            end

            local service = uri:match("^kuma://kuma.io/service/(.+)$")
            if service ~= nil and service ~= '' then
                request_handle:headers():add("X-Kuma-Forwarded-Client-Service", service)
            end

            local zone = uri:match("^kuma://kuma.io/zone/(.+)$")
            if zone ~= nil and zone ~= '' then
                request_handle:headers():add("X-Kuma-Forwarded-Client-Zone", zone)
            end
        end
    end
end

function envoy_on_response(handle)
end

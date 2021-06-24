function envoy_on_request(request_handle)
    xfcc = request_handle:headers():get("x-forwarded-client-cert")

    if xfcc == nil or xfcc == '' then
        return
    end

    spiffe = nil
    service = nil

    for str in string.gmatch(xfcc, "([^;]+)") do
        uri_match_result = str:match("URI=(%S+)")
        if uri_match_result ~= nil then
            spiffe = uri_match_result
            mesh = spiffe:match("spiffe://(%S+)/")
            service = spiffe:match("spiffe://" .. mesh .. "/(%S+)")
        end
    end

    if spiffe ~= nil then
        request_handle:headers():add("X-Kuma-Forwarded-Client-Cert", spiffe)
    end

    if service ~= nil then
        request_handle:headers():add("X-Kuma-Forwarded-Client-Service", service)
    end
end

function envoy_on_response(handle)
end

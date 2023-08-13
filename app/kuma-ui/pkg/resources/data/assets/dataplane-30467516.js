function p(n){let o=[];n.networking.inbound&&(o=n.networking.inbound.filter(e=>"tags"in e).flatMap(e=>Object.entries(e.tags)).map(([e,i])=>`${e}=${i}`)),n.networking.gateway&&(o=Object.entries(n.networking.gateway.tags).map(([e,i])=>`${e}=${i}`));const t=Array.from(new Set(o));return t.sort((e,i)=>e.localeCompare(i)),t.map(e=>e.split("=")).map(([e,i])=>({label:e,value:i}))}function r(n={subscriptions:[]}){return(n.subscriptions??[]).some(t=>{var e;return((e=t.connectTime)==null?void 0:e.length)&&!t.disconnectTime})?"online":"offline"}function l(n,o={subscriptions:[]}){const t=n.networking.inbound??[],e=t.filter(s=>s.health&&!s.health.ready).map(s=>`Inbound on port ${s.port} is not ready (kuma.io/service: ${s.tags["kuma.io/service"]})`);let i;switch(!0){case t.length===0:i="online";break;case e.length===t.length:i="offline";break;case e.length>0:i="partially_degraded";break;default:i=r(o)}return{status:i,reason:e}}function m(n){if(n===void 0||n.subscriptions.length===0)return null;const o={},t=n.subscriptions[n.subscriptions.length-1];return t.version===void 0?null:(t.version.envoy&&(o.envoy=t.version.envoy.version),t.version.kumaDp&&(o.kumaDp=t.version.kumaDp.version),t.version.dependencies&&Object.entries(t.version.dependencies).forEach(([e,i])=>{o[e]=i}),o)}function O(n,o){if(n.dataplaneInsight===void 0||n.dataplaneInsight.mTLS===void 0)return null;const{mTLS:t}=n.dataplaneInsight,e=new Date(t.certificateExpirationTime),i=new Date(e.getTime()+e.getTimezoneOffset()*6e4);return{certificateExpirationTime:o(i.toISOString()),lastCertificateRegeneration:o(t.lastCertificateRegeneration),certificateRegenerations:t.certificateRegenerations}}function _(n){var e,i;return((e=n.kumaDp)==null?void 0:e.kumaCpCompatible)??!0?((i=n.envoy)==null?void 0:i.kumaDpCompatible)??!0?{kind:a}:{kind:u,payload:{envoy:n.envoy.version,kumaDp:n.kumaDp.version}}:{kind:c,payload:{kumaDp:n.kumaDp.version}}}const a="COMPATIBLE",f="INCOMPATIBLE_ZONE_CP_AND_KUMA_DP_VERSIONS",g="INCOMPATIBLE_ZONE_AND_GLOBAL_CPS_VERSIONS",c="INCOMPATIBLE_UNSUPPORTED_KUMA_DP",u="INCOMPATIBLE_UNSUPPORTED_ENVOY",E="INCOMPATIBLE_WRONG_FORMAT";export{a as C,E as I,l as a,m as b,_ as c,p as d,f as e,g as f,r as g,c as h,u as i,O as p};

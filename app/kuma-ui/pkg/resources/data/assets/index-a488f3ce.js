import{ah as s}from"./index-0d828147.js";function c(e){var t;const n=((t=e.zoneInsight)==null?void 0:t.subscriptions)??[];if(n.length>0){const o=n[n.length-1];if(o.config){const i=JSON.parse(o.config);return s(i,"dpServer.auth.type","")}}return""}function f(e){var o,i;if(e.zone.enabled===!1)return"disabled";const n=((o=e.zoneInsight)==null?void 0:o.subscriptions)??[];if(n.length===0)return"offline";const t=n[n.length-1];return(i=t.connectTime)!=null&&i.length&&!t.disconnectTime?"online":"offline"}function u(e){var n;for(const t of((n=e.zoneInsight)==null?void 0:n.subscriptions)??[])if(t.config)return JSON.parse(t.config).environment;return""}export{u as a,c as b,f as g};

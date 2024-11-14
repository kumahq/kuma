import{d as R,e as p,o as d,k as z,w as t,a,j as w,x,Q as B,b as n,i as m,P as k,t as o,S as K,n as O,a0 as E,K as M,l as y,c as _,A as H,F as I,G as b,C as U,m as Z}from"./index-CUmbT3FY.js";import{q as G}from"./kong-icons.es678-D9cWiXFe.js";import{S as P}from"./SummaryView-mq1hk7FF.js";const $=["data-testid","innerHTML"],F={"data-testid":"detail-view-details",class:"stack"},X={class:"columns"},j=["innerHTML"],q={key:0},Q=R({__name:"ZoneDetailView",props:{data:{}},setup(g){const i=g;return(J,W)=>{const v=p("KTooltip"),V=p("KCard"),S=p("XAction"),T=p("RouterView"),A=p("AppView"),D=p("DataSource"),N=p("RouteView");return d(),z(N,{name:"zone-cp-detail-view",params:{zone:"",subscription:""}},{default:t(({t:s,uri:L,route:h,me:u})=>{var f,C;return[a(D,{src:L(w(x),"/control-plane/outdated/:version",{version:((C=(f=i.data.zoneInsight.version)==null?void 0:f.kumaCp)==null?void 0:C.version)??"-"})},{default:t(({data:r})=>[a(A,{docs:s("zones.href.docs.cta")},B({default:t(()=>[n(),m("div",F,[a(V,null,{default:t(()=>[m("div",X,[a(k,null,{title:t(()=>[n(o(s("http.api.property.status")),1)]),body:t(()=>[a(K,{status:i.data.state},null,8,["status"])]),_:2},1024),n(),a(k,{class:O({version:!0,outdated:r==null?void 0:r.outdated})},{title:t(()=>[n(o(s("zone-cps.routes.item.version"))+" ",1),(r==null?void 0:r.outdated)===!0?(d(),z(v,{key:0,"max-width":"300"},{content:t(()=>[m("div",{innerHTML:s("zone-cps.routes.item.version_warning")},null,8,j)]),default:t(()=>[a(w(G),{color:w(E),size:w(M)},null,8,["color","size"]),n()]),_:2},1024)):y("",!0)]),body:t(()=>{var e,c;return[n(o(((c=(e=i.data.zoneInsight.version)==null?void 0:e.kumaCp)==null?void 0:c.version)??"—"),1)]}),_:2},1032,["class"]),n(),a(k,null,{title:t(()=>[n(o(s("http.api.property.type")),1)]),body:t(()=>[n(o(s(`common.product.environment.${i.data.zoneInsight.environment||"unknown"}`)),1)]),_:2},1024),n(),a(k,null,{title:t(()=>[n(o(s("zone-cps.routes.item.authentication_type")),1)]),body:t(()=>[n(o(i.data.zoneInsight.authenticationType||s("common.not_applicable")),1)]),_:2},1024)])]),_:2},1024),n(),i.data.zoneInsight.subscriptions.length>0?(d(),_("div",q,[m("h2",null,o(s("zone-cps.detail.subscriptions")),1),n(),a(H,{headers:[{...u.get("headers.zoneInstanceId"),label:s("zone-cps.routes.items.headers.zoneInstanceId"),key:"zoneInstanceId"},{...u.get("headers.version"),label:s("zone-cps.routes.items.headers.version"),key:"version"},{...u.get("headers.connected"),label:s("zone-cps.routes.items.headers.connected"),key:"connected"},{...u.get("headers.disconnected"),label:s("zone-cps.routes.items.headers.disconnected"),key:"disconnected"},{...u.get("headers.responses"),label:s("zone-cps.routes.items.headers.responses"),key:"responses"}],"is-selected-row":e=>e.id===h.params.subscription,items:i.data.zoneInsight.subscriptions.map((e,c,l)=>l[l.length-(c+1)]),onResize:u.set},{zoneInstanceId:t(({row:e})=>[a(S,{"data-action":"",to:{name:"zone-cp-subscription-summary-view",params:{subscription:e.id}}},{default:t(()=>[n(o(e.zoneInstanceId),1)]),_:2},1032,["to"])]),version:t(({row:e})=>{var c,l;return[n(o(((l=(c=e.version)==null?void 0:c.kumaCp)==null?void 0:l.version)??"-"),1)]}),connected:t(({row:e})=>[n(o(s("common.formats.datetime",{value:Date.parse(e.connectTime??"")})),1)]),disconnected:t(({row:e})=>[e.disconnectTime?(d(),_(I,{key:0},[n(o(s("common.formats.datetime",{value:Date.parse(e.disconnectTime)})),1)],64)):y("",!0)]),responses:t(({row:e})=>{var c;return[(d(!0),_(I,null,b([((c=e.status)==null?void 0:c.total)??{}],l=>(d(),_(I,null,[n(o(l.responsesSent)+"/"+o(l.responsesAcknowledged),1)],64))),256))]}),_:2},1032,["headers","is-selected-row","items","onResize"]),n(),a(T,null,{default:t(({Component:e})=>[h.child()?(d(),z(P,{key:0,width:"670px",onClose:function(){h.replace({name:"zone-cp-detail-view",params:{zone:h.params.zone}})}},{default:t(()=>[(d(),z(U(e),{data:i.data.zoneInsight.subscriptions},{default:t(()=>[m("p",null,o(s("zone-cps.routes.item.subscription_intro")),1)]),_:2},1032,["data"]))]),_:2},1032,["onClose"])):y("",!0)]),_:2},1024)])):y("",!0)])]),_:2},[i.data.warnings.length>0?{name:"notifications",fn:t(()=>[m("ul",null,[(d(!0),_(I,null,b(i.data.warnings,e=>(d(),_("li",{key:e.kind,"data-testid":`warning-${e.kind}`,innerHTML:s(`common.warnings.${e.kind}`,{...e.payload,...e.kind==="INCOMPATIBLE_ZONE_AND_GLOBAL_CPS_VERSIONS"?{globalCpVersion:(r==null?void 0:r.version)??""}:{}})},null,8,$))),128))])]),key:"0"}:void 0]),1032,["docs"])]),_:2},1032,["src"])]}),_:1})}}}),ne=Z(Q,[["__scopeId","data-v-ec428780"]]);export{ne as default};

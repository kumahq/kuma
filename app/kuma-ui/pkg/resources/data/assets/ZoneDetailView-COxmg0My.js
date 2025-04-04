import{d as G,m as z,w as e,b as n,r as p,c as _,e as o,v as B,F as f,p as b,L as g,S as K,t as d,n as Z,C as H,K as J,q as D,s as M,o as l,B as P}from"./index-q-proItx.js";import{S as Q}from"./SummaryView-DQLVqYO3.js";const ee=G({__name:"ZoneDetailView",props:{data:{}},setup(R){const i=R;return(U,s)=>{const h=p("XI18n"),T=p("XNotification"),w=p("XBadge"),N=p("XIcon"),I=p("XLayout"),x=p("XAboutCard"),$=p("XAction"),L=p("RouterView"),j=p("XCard"),E=p("AppView"),F=p("DataSource"),O=p("RouteView");return l(),z(O,{name:"zone-cp-detail-view",params:{zone:"",subscription:""}},{default:e(({t:a,uri:q,route:y,me:m})=>{var C,k;return[n(F,{src:q(M(P),"/control-plane/outdated/:version",{version:((k=(C=i.data.zoneInsight.version)==null?void 0:C.kumaCp)==null?void 0:k.version)??"-"})},{default:e(({data:c})=>[n(E,{docs:a("zones.href.docs.cta"),notifications:!0},{default:e(()=>{var X,v,V,S,A;return[(l(!0),_(f,null,B([{bool:i.data.zoneInsight.store==="memory",key:"store-memory"},{bool:!((v=(X=i.data.zoneInsight.version)==null?void 0:X.kumaCp)!=null&&v.kumaCpGlobalCompatible),key:"global-cp-incompatible",params:{zoneCpVersion:((S=(V=i.data.zoneInsight.version)==null?void 0:V.kumaCp)==null?void 0:S.version)??"-",globalCpVersion:(c==null?void 0:c.version)??""}},{bool:(((A=i.data.zoneInsight.connectedSubscription)==null?void 0:A.status.total.responsesRejected)??0)>0,key:"global-nack-response"}],({bool:t,key:r,params:u})=>(l(),_(f,{key:r},[t?(l(),z(T,{key:0,"data-testid":`warning-${r}`,uri:`zone-cps.notifications.${r}.${i.data.id}`},{default:e(()=>[n(h,{path:`zone-cps.notifications.${r}`,params:Object.fromEntries(Object.entries(u??{}))},null,8,["path","params"])]),_:2},1032,["data-testid","uri"])):b("",!0)],64))),128)),s[15]||(s[15]=o()),n(I,{"data-testid":"detail-view-details",type:"stack"},{default:e(()=>[n(x,{title:a("zone-cps.detail.about.title"),created:i.data.creationTime,modified:i.data.modificationTime},{default:e(()=>[n(g,{layout:"horizontal"},{title:e(()=>[o(d(a("http.api.property.status")),1)]),body:e(()=>[n(K,{status:i.data.state},null,8,["status"])]),_:2},1024),s[5]||(s[5]=o()),n(g,{layout:"horizontal",class:Z({version:!0,outdated:c==null?void 0:c.outdated})},{title:e(()=>[o(d(a("zone-cps.routes.item.version")),1)]),body:e(()=>[n(I,{type:"separated"},{default:e(()=>[n(w,{appearance:(c==null?void 0:c.outdated)===!0?"warning":"decorative"},{default:e(()=>{var t,r;return[o(d(((r=(t=i.data.zoneInsight.version)==null?void 0:t.kumaCp)==null?void 0:r.version)??"—"),1)]}),_:2},1032,["appearance"]),s[1]||(s[1]=o()),(c==null?void 0:c.outdated)===!0?(l(),z(N,{key:0,name:"info"},{default:e(()=>[n(h,{path:"zone-cps.routes.item.version_warning"})]),_:1})):b("",!0)]),_:2},1024)]),_:2},1032,["class"]),s[6]||(s[6]=o()),n(g,{layout:"horizontal"},{title:e(()=>[o(d(a("http.api.property.type")),1)]),body:e(()=>[n(w,{appearance:"decorative"},{default:e(()=>[o(d(a(`common.product.environment.${i.data.zoneInsight.environment||"unknown"}`)),1)]),_:2},1024)]),_:2},1024),s[7]||(s[7]=o()),n(g,{layout:"horizontal"},{title:e(()=>[o(d(a("zone-cps.routes.item.authentication_type")),1)]),body:e(()=>[n(w,{appearance:"decorative"},{default:e(()=>[o(d(i.data.zoneInsight.authenticationType||a("common.not_applicable")),1)]),_:2},1024)]),_:2},1024)]),_:2},1032,["title","created","modified"]),s[14]||(s[14]=o()),i.data.zoneInsight.subscriptions.length>0?(l(),z(j,{key:0},{title:e(()=>[D("h2",null,d(a("zone-cps.detail.subscriptions")),1)]),default:e(()=>[s[12]||(s[12]=o()),n(H,{headers:[{...m.get("headers.zoneInstanceId"),label:a("zone-cps.routes.items.headers.zoneInstanceId"),key:"zoneInstanceId"},{...m.get("headers.version"),label:a("zone-cps.routes.items.headers.version"),key:"version"},{...m.get("headers.connected"),label:a("zone-cps.routes.items.headers.connected"),key:"connected"},{...m.get("headers.disconnected"),label:a("zone-cps.routes.items.headers.disconnected"),key:"disconnected"},{...m.get("headers.responses"),label:a("zone-cps.routes.items.headers.responses"),key:"responses"}],"is-selected-row":t=>t.id===y.params.subscription,items:i.data.zoneInsight.subscriptions.map((t,r,u)=>u[u.length-(r+1)]),onResize:m.set},{zoneInstanceId:e(({row:t})=>[n($,{"data-action":"",to:{name:"zone-cp-subscription-summary-view",params:{subscription:t.id}}},{default:e(()=>[o(d(t.zoneInstanceId),1)]),_:2},1032,["to"])]),version:e(({row:t})=>{var r,u;return[o(d(((u=(r=t.version)==null?void 0:r.kumaCp)==null?void 0:u.version)??"-"),1)]}),connected:e(({row:t})=>[o(d(a("common.formats.datetime",{value:Date.parse(t.connectTime??"")})),1)]),disconnected:e(({row:t})=>[t.disconnectTime?(l(),_(f,{key:0},[o(d(a("common.formats.datetime",{value:Date.parse(t.disconnectTime)})),1)],64)):b("",!0)]),responses:e(({row:t})=>{var r;return[(l(!0),_(f,null,B([((r=t.status)==null?void 0:r.total)??{}],u=>(l(),_(f,null,[o(d(u.responsesSent)+"/"+d(u.responsesAcknowledged),1)],64))),256))]}),_:2},1032,["headers","is-selected-row","items","onResize"]),s[13]||(s[13]=o()),n(L,null,{default:e(({Component:t})=>[y.child()?(l(),z(Q,{key:0,width:"670px",onClose:function(){y.replace({name:"zone-cp-detail-view",params:{zone:y.params.zone}})}},{default:e(()=>[(l(),z(J(t),{data:i.data.zoneInsight.subscriptions},{default:e(()=>[D("p",null,d(a("zone-cps.routes.item.subscription_intro")),1)]),_:2},1032,["data"]))]),_:2},1032,["onClose"])):b("",!0)]),_:2},1024)]),_:2},1024)):b("",!0)]),_:2},1024)]}),_:2},1032,["docs"])]),_:2},1032,["src"])]}),_:1})}}});export{ee as default};

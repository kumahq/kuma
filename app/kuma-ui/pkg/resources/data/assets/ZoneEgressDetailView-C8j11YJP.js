import{d as G,J,r as l,o as r,m,w as e,b as t,U as A,e as n,t as i,S as K,q as w,c as _,F as g,p as V,$ as P,v as C,K as U,s as H,C as M,_ as Q}from"./index-C-Llvxgw.js";import{S as W}from"./SummaryView-CzmbSpU2.js";import{C as S,b as B,a as Y}from"./ConnectionTraffic-BieQZoHt.js";import"./TagList-CVIOP13G.js";const ee=G({__name:"ZoneEgressDetailView",props:{data:{}},setup(T){const d=T,D=J();return(te,o)=>{const R=l("XBadge"),$=l("XCopyButton"),L=l("XAboutCard"),z=l("XAction"),X=l("DataCollection"),h=l("XLayout"),q=l("XInputSwitch"),I=l("XCard"),N=l("RouterView"),O=l("DataLoader"),j=l("AppView"),F=l("RouteView");return r(),m(F,{name:"zone-egress-detail-view",params:{inactive:Boolean,subscription:"",proxy:""}},{default:e(({t:p,route:u,me:f,uri:Z})=>[t(j,null,{default:e(()=>[t(h,{type:"stack"},{default:e(()=>[t(L,{title:p("zone-egresses.routes.item.about.title"),created:d.data.creationTime,modified:d.data.modificationTime},{default:e(()=>[t(A,{layout:"horizontal"},{title:e(()=>[n(i(p("http.api.property.status")),1)]),body:e(()=>[t(K,{status:d.data.state},null,8,["status"])]),_:2},1024),o[3]||(o[3]=n()),d.data.namespace.length>0?(r(),m(A,{key:0,layout:"horizontal"},{title:e(()=>[n(i(p("http.api.property.namespace")),1)]),body:e(()=>[t(R,{appearance:"decorative"},{default:e(()=>[n(i(d.data.namespace),1)]),_:1})]),_:2},1024)):w("",!0),o[4]||(o[4]=n()),t(A,{layout:"horizontal"},{title:e(()=>[n(i(p("http.api.property.address")),1)]),body:e(()=>[d.data.zoneEgress.socketAddress.length>0?(r(),m($,{key:0,variant:"badge",format:"default",text:d.data.zoneEgress.socketAddress},null,8,["text"])):(r(),_(g,{key:1},[n(i(p("common.detail.none")),1)],64))]),_:2},1024)]),_:2},1032,["title","created","modified"]),o[16]||(o[16]=n()),t(O,{src:Z(V(P),"/connections/stats/for/:proxyType/:name/:mesh/:socketAddress",{name:u.params.proxy,mesh:"*",socketAddress:d.data.zoneEgress.socketAddress,proxyType:"zone-egress"})},{default:e(({data:s,refresh:c})=>[t(I,null,{default:e(()=>[t(h,{type:"columns"},{default:e(()=>[t(S,null,{default:e(()=>[t(h,{type:"stack",size:"small"},{default:e(()=>[t(X,{type:"inbounds",items:Object.entries(s.inbounds)},{default:e(({items:a})=>[(r(!0),_(g,null,C(a,([k,v])=>(r(),m(B,{key:`${k}`,protocol:"",traffic:v},{default:e(()=>[t(z,{"data-action":"",to:{name:(y=>y.includes("bound")?y.replace("-outbound-","-inbound-"):"zone-egress-connection-inbound-summary-stats-view")(String(V(D).name)),params:{connection:k},query:{inactive:u.params.inactive}}},{default:e(()=>[n(`
                          :`+i(k.split("_").at(-1)),1)]),_:2},1032,["to"])]),_:2},1032,["traffic"]))),128))]),_:2},1032,["items"])]),_:2},1024)]),_:2},1024),o[9]||(o[9]=n()),t(S,null,{actions:e(()=>[t(q,{checked:u.params.inactive,"data-testid":"dataplane-outbounds-inactive-toggle",onChange:a=>u.update({inactive:a})},{label:e(()=>o[5]||(o[5]=[n(`
                      Show inactive
                    `)])),_:2},1032,["checked","onChange"]),o[7]||(o[7]=n()),t(z,{action:"refresh",appearance:"primary",onClick:c},{default:e(()=>o[6]||(o[6]=[n(`
                    Refresh
                  `)])),_:2},1032,["onClick"])]),default:e(()=>[o[8]||(o[8]=n()),(r(),_(g,null,C(["upstream"],a=>t(X,{key:a,type:"outbounds",items:Object.entries(s.outbounds)},{default:e(({items:k})=>[t(X,{type:"activeOutbounds",predicate:u.params.inactive?void 0:([v,y])=>{var b,x;return((typeof y.tcp<"u"?(b=y.tcp)==null?void 0:b[`${a}_cx_rx_bytes_total`]:(x=y.http)==null?void 0:x[`${a}_rq_total`])??0)>0},items:k},{default:e(({items:v})=>[v.length>0?(r(),m(Y,{key:0,type:"outbound"},{default:e(()=>[(r(),_(g,null,C([/-([a-f0-9]){16}$/],y=>t(h,{key:y,type:"stack",size:"small"},{default:e(()=>[(r(!0),_(g,null,C(v,([b,x])=>(r(),m(B,{key:`${b}`,"data-testid":"dataplane-outbound",protocol:"",traffic:x,direction:a},{default:e(()=>[t(z,{"data-action":"",to:{name:(E=>E.includes("bound")?E.replace("-inbound-","-outbound-"):"zone-egress-connection-outbound-summary-stats-view")(String(V(D).name)),params:{connection:b},query:{inactive:u.params.inactive}}},{default:e(()=>[n(i(b),1)]),_:2},1032,["to"])]),_:2},1032,["traffic","direction"]))),128))]),_:2},1024)),64))]),_:2},1024)):w("",!0)]),_:2},1032,["predicate","items"])]),_:2},1032,["items"])),64))]),_:2},1024)]),_:2},1024)]),_:2},1024),o[10]||(o[10]=n()),t(N,null,{default:e(a=>[a.route.name!==u.name?(r(),m(W,{key:0,width:"670px",onClose:function(){u.replace({name:"zone-egress-detail-view",params:{proxyType:"egresses",proxy:u.params.proxy}})}},{default:e(()=>[(r(),m(U(a.Component),{data:u.params.subscription.length>0?d.data.zoneEgressInsight.subscriptions:a.route.name.includes("-inbound-")?[d.data.zoneEgress]:(s==null?void 0:s.outbounds)||{},networking:d.data.zoneEgress.networking},null,8,["data","networking"]))]),_:2},1032,["onClose"])):w("",!0)]),_:2},1024)]),_:2},1032,["src"]),o[17]||(o[17]=n()),d.data.zoneEgressInsight.subscriptions.length>0?(r(),m(I,{key:0},{title:e(()=>[H("h2",null,i(p("zone-egresses.routes.item.subscriptions.title")),1)]),default:e(()=>[o[15]||(o[15]=n()),t(M,{headers:[{...f.get("headers.instanceId"),label:p("http.api.property.instanceId"),key:"instanceId"},{...f.get("headers.version"),label:p("http.api.property.version"),key:"version"},{...f.get("headers.connected"),label:p("http.api.property.connected"),key:"connected"},{...f.get("headers.disconnected"),label:p("http.api.property.disconnected"),key:"disconnected"},{...f.get("headers.responses"),label:p("http.api.property.responses"),key:"responses"}],"is-selected-row":s=>s.id===u.params.subscription,items:d.data.zoneEgressInsight.subscriptions.map((s,c,a)=>a[a.length-(c+1)]),onResize:f.set},{instanceId:e(({row:s})=>[t(z,{"data-action":"",to:{name:"zone-egress-subscription-summary-view",params:{subscription:s.id}}},{default:e(()=>[n(i(s.controlPlaneInstanceId),1)]),_:2},1032,["to"])]),version:e(({row:s})=>{var c,a;return[n(i(((a=(c=s.version)==null?void 0:c.kumaDp)==null?void 0:a.version)??"-"),1)]}),connected:e(({row:s})=>[n(i(p("common.formats.datetime",{value:Date.parse(s.connectTime??"")})),1)]),disconnected:e(({row:s})=>[s.disconnectTime?(r(),_(g,{key:0},[n(i(p("common.formats.datetime",{value:Date.parse(s.disconnectTime)})),1)],64)):w("",!0)]),responses:e(({row:s})=>{var c;return[(r(!0),_(g,null,C([((c=s.status)==null?void 0:c.total)??{}],a=>(r(),_(g,null,[n(i(a.responsesSent)+"/"+i(a.responsesAcknowledged),1)],64))),256))]}),_:2},1032,["headers","is-selected-row","items","onResize"])]),_:2},1024)):w("",!0)]),_:2},1024)]),_:2},1024)]),_:1})}}}),re=Q(ee,[["__scopeId","data-v-47356dbe"]]);export{re as default};

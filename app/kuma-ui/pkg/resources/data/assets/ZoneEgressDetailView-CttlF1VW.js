import{d as I,e as l,o as i,p as u,w as e,a as d,l as k,b as s,Q as f,t as a,S as T,q as _,c as y,J as w,A as X,K as x,F as B}from"./index-DH1Ug2X_.js";import{S as R}from"./SummaryView-pv74nbtW.js";const S={class:"stack"},N={class:"columns"},K={key:0},J=I({__name:"ZoneEgressDetailView",props:{data:{}},setup(v){const r=v;return(L,n)=>{const b=l("XTimespan"),h=l("XCopyButton"),z=l("XLayout"),C=l("KCard"),V=l("XAction"),E=l("RouterView"),A=l("AppView"),D=l("RouteView");return i(),u(D,{name:"zone-egress-detail-view",params:{subscription:"",zoneEgress:""}},{default:e(({t:o,route:g,me:m})=>[d(A,null,{default:e(()=>[k("div",S,[d(C,null,{default:e(()=>[d(z,{type:"stack"},{default:e(()=>[d(b,{start:o("common.formats.datetime",{value:Date.parse(r.data.creationTime)}),end:o("common.formats.datetime",{value:Date.parse(r.data.modificationTime)})},null,8,["start","end"]),n[6]||(n[6]=s()),k("div",N,[d(f,null,{title:e(()=>[s(a(o("http.api.property.status")),1)]),body:e(()=>[d(T,{status:r.data.state},null,8,["status"])]),_:2},1024),n[4]||(n[4]=s()),r.data.namespace.length>0?(i(),u(f,{key:0},{title:e(()=>n[1]||(n[1]=[s(`
                  Namespace
                `)])),body:e(()=>[s(a(r.data.namespace),1)]),_:1})):_("",!0),n[5]||(n[5]=s()),d(f,null,{title:e(()=>[s(a(o("http.api.property.address")),1)]),body:e(()=>[r.data.zoneEgress.socketAddress.length>0?(i(),u(h,{key:0,text:r.data.zoneEgress.socketAddress},null,8,["text"])):(i(),y(w,{key:1},[s(a(o("common.detail.none")),1)],64))]),_:2},1024)])]),_:2},1024)]),_:2},1024),n[13]||(n[13]=s()),r.data.zoneEgressInsight.subscriptions.length>0?(i(),y("div",K,[k("h2",null,a(o("zone-egresses.routes.item.subscriptions.title")),1),n[11]||(n[11]=s()),d(X,{headers:[{...m.get("headers.instanceId"),label:o("http.api.property.instanceId"),key:"instanceId"},{...m.get("headers.version"),label:o("http.api.property.version"),key:"version"},{...m.get("headers.connected"),label:o("http.api.property.connected"),key:"connected"},{...m.get("headers.disconnected"),label:o("http.api.property.disconnected"),key:"disconnected"},{...m.get("headers.responses"),label:o("http.api.property.responses"),key:"responses"}],"is-selected-row":t=>t.id===g.params.subscription,items:r.data.zoneEgressInsight.subscriptions.map((t,c,p)=>p[p.length-(c+1)]),onResize:m.set},{instanceId:e(({row:t})=>[d(V,{"data-action":"",to:{name:"zone-egress-subscription-summary-view",params:{subscription:t.id}}},{default:e(()=>[s(a(t.controlPlaneInstanceId),1)]),_:2},1032,["to"])]),version:e(({row:t})=>{var c,p;return[s(a(((p=(c=t.version)==null?void 0:c.kumaDp)==null?void 0:p.version)??"-"),1)]}),connected:e(({row:t})=>[s(a(o("common.formats.datetime",{value:Date.parse(t.connectTime??"")})),1)]),disconnected:e(({row:t})=>[t.disconnectTime?(i(),y(w,{key:0},[s(a(o("common.formats.datetime",{value:Date.parse(t.disconnectTime)})),1)],64)):_("",!0)]),responses:e(({row:t})=>{var c;return[(i(!0),y(w,null,x([((c=t.status)==null?void 0:c.total)??{}],p=>(i(),y(w,null,[s(a(p.responsesSent)+"/"+a(p.responsesAcknowledged),1)],64))),256))]}),_:2},1032,["headers","is-selected-row","items","onResize"]),n[12]||(n[12]=s()),d(E,null,{default:e(({Component:t})=>[g.child()?(i(),u(R,{key:0,width:"670px",onClose:function(){g.replace({name:"zone-egress-detail-view",params:{zoneEgress:g.params.zoneEgress}})}},{default:e(()=>[(i(),u(B(t),{data:r.data.zoneEgressInsight.subscriptions},null,8,["data"]))]),_:2},1032,["onClose"])):_("",!0)]),_:2},1024)])):_("",!0)])]),_:2},1024)]),_:1})}}});export{J as default};

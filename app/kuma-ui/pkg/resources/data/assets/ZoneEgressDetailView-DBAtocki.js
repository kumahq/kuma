import{d as A,e as u,o as r,m,w as e,a as p,k,Q as b,b as t,t as a,S as I,p as _,c as g,J as w,A as x,K as B,F as D}from"./index-CKcsX_-l.js";import{S as R}from"./SummaryView-BIwsKbzL.js";const S={class:"stack"},N={class:"columns"},T={key:0},J=A({__name:"ZoneEgressDetailView",props:{data:{}},setup(h){const i=h;return(X,n)=>{const v=u("XCopyButton"),f=u("KCard"),z=u("XAction"),C=u("RouterView"),V=u("AppView"),E=u("RouteView");return r(),m(E,{name:"zone-egress-detail-view",params:{subscription:"",zoneEgress:""}},{default:e(({t:o,route:y,me:c})=>[p(V,null,{default:e(()=>[k("div",S,[p(f,null,{default:e(()=>[k("div",N,[p(b,null,{title:e(()=>[t(a(o("http.api.property.status")),1)]),body:e(()=>[p(I,{status:i.data.state},null,8,["status"])]),_:2},1024),n[4]||(n[4]=t()),i.data.namespace.length>0?(r(),m(b,{key:0},{title:e(()=>n[1]||(n[1]=[t(`
                Namespace
              `)])),body:e(()=>[t(a(i.data.namespace),1)]),_:1})):_("",!0),n[5]||(n[5]=t()),p(b,null,{title:e(()=>[t(a(o("http.api.property.address")),1)]),body:e(()=>[i.data.zoneEgress.socketAddress.length>0?(r(),m(v,{key:0,text:i.data.zoneEgress.socketAddress},null,8,["text"])):(r(),g(w,{key:1},[t(a(o("common.detail.none")),1)],64))]),_:2},1024)])]),_:2},1024),n[12]||(n[12]=t()),i.data.zoneEgressInsight.subscriptions.length>0?(r(),g("div",T,[k("h2",null,a(o("zone-egresses.routes.item.subscriptions.title")),1),n[10]||(n[10]=t()),p(x,{headers:[{...c.get("headers.instanceId"),label:o("http.api.property.instanceId"),key:"instanceId"},{...c.get("headers.version"),label:o("http.api.property.version"),key:"version"},{...c.get("headers.connected"),label:o("http.api.property.connected"),key:"connected"},{...c.get("headers.disconnected"),label:o("http.api.property.disconnected"),key:"disconnected"},{...c.get("headers.responses"),label:o("http.api.property.responses"),key:"responses"}],"is-selected-row":s=>s.id===y.params.subscription,items:i.data.zoneEgressInsight.subscriptions.map((s,l,d)=>d[d.length-(l+1)]),onResize:c.set},{instanceId:e(({row:s})=>[p(z,{"data-action":"",to:{name:"zone-egress-subscription-summary-view",params:{subscription:s.id}}},{default:e(()=>[t(a(s.controlPlaneInstanceId),1)]),_:2},1032,["to"])]),version:e(({row:s})=>{var l,d;return[t(a(((d=(l=s.version)==null?void 0:l.kumaDp)==null?void 0:d.version)??"-"),1)]}),connected:e(({row:s})=>[t(a(o("common.formats.datetime",{value:Date.parse(s.connectTime??"")})),1)]),disconnected:e(({row:s})=>[s.disconnectTime?(r(),g(w,{key:0},[t(a(o("common.formats.datetime",{value:Date.parse(s.disconnectTime)})),1)],64)):_("",!0)]),responses:e(({row:s})=>{var l;return[(r(!0),g(w,null,B([((l=s.status)==null?void 0:l.total)??{}],d=>(r(),g(w,null,[t(a(d.responsesSent)+"/"+a(d.responsesAcknowledged),1)],64))),256))]}),_:2},1032,["headers","is-selected-row","items","onResize"]),n[11]||(n[11]=t()),p(C,null,{default:e(({Component:s})=>[y.child()?(r(),m(R,{key:0,width:"670px",onClose:function(){y.replace({name:"zone-egress-detail-view",params:{zoneEgress:y.params.zoneEgress}})}},{default:e(()=>[(r(),m(D(s),{data:i.data.zoneEgressInsight.subscriptions},null,8,["data"]))]),_:2},1032,["onClose"])):_("",!0)]),_:2},1024)])):_("",!0)])]),_:2},1024)]),_:1})}}});export{J as default};

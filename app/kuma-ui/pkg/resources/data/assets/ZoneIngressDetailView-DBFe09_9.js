import{d as V,e as u,o as r,m,w as e,a as d,k as I,P as w,b as t,t as a,S as x,p as _,c as g,H as y,A as S,J as B,E as D}from"./index-CjjKwNo4.js";import{S as R}from"./SummaryView-BhSKxYbl.js";const N={class:"stack"},T={class:"columns"},X={key:0},F=V({__name:"ZoneIngressDetailView",props:{data:{}},setup(b){const i=b;return(E,n)=>{const v=u("XCopyButton"),h=u("KCard"),z=u("XAction"),f=u("RouterView"),C=u("AppView"),A=u("RouteView");return r(),m(A,{name:"zone-ingress-detail-view",params:{subscription:"",zoneIngress:""}},{default:e(({t:o,me:c,route:k})=>[d(C,null,{default:e(()=>[I("div",N,[d(h,null,{default:e(()=>[I("div",T,[d(w,null,{title:e(()=>[t(a(o("http.api.property.status")),1)]),body:e(()=>[d(x,{status:i.data.state},null,8,["status"])]),_:2},1024),n[5]||(n[5]=t()),i.data.namespace.length>0?(r(),m(w,{key:0},{title:e(()=>n[1]||(n[1]=[t(`
                Namespace
              `)])),body:e(()=>[t(a(i.data.namespace),1)]),_:1})):_("",!0),n[6]||(n[6]=t()),d(w,null,{title:e(()=>[t(a(o("http.api.property.address")),1)]),body:e(()=>[i.data.zoneIngress.socketAddress.length>0?(r(),m(v,{key:0,text:i.data.zoneIngress.socketAddress},null,8,["text"])):(r(),g(y,{key:1},[t(a(o("common.detail.none")),1)],64))]),_:2},1024),n[7]||(n[7]=t()),d(w,null,{title:e(()=>[t(a(o("http.api.property.advertisedAddress")),1)]),body:e(()=>[i.data.zoneIngress.advertisedSocketAddress.length>0?(r(),m(v,{key:0,text:i.data.zoneIngress.advertisedSocketAddress},null,8,["text"])):(r(),g(y,{key:1},[t(a(o("common.detail.none")),1)],64))]),_:2},1024)])]),_:2},1024),n[14]||(n[14]=t()),i.data.zoneIngressInsight.subscriptions.length>0?(r(),g("div",X,[I("h2",null,a(o("zone-ingresses.routes.item.subscriptions.title")),1),n[12]||(n[12]=t()),d(S,{headers:[{...c.get("headers.instanceId"),label:o("http.api.property.instanceId"),key:"instanceId"},{...c.get("headers.version"),label:o("http.api.property.version"),key:"version"},{...c.get("headers.connected"),label:o("http.api.property.connected"),key:"connected"},{...c.get("headers.disconnected"),label:o("http.api.property.disconnected"),key:"disconnected"},{...c.get("headers.responses"),label:o("http.api.property.responses"),key:"responses"}],"is-selected-row":s=>s.id===k.params.subscription,items:i.data.zoneIngressInsight.subscriptions.map((s,l,p)=>p[p.length-(l+1)]),onResize:c.set},{instanceId:e(({row:s})=>[d(z,{"data-action":"",to:{name:"zone-ingress-subscription-summary-view",params:{subscription:s.id}}},{default:e(()=>[t(a(s.controlPlaneInstanceId),1)]),_:2},1032,["to"])]),version:e(({row:s})=>{var l,p;return[t(a(((p=(l=s.version)==null?void 0:l.kumaDp)==null?void 0:p.version)??"-"),1)]}),connected:e(({row:s})=>[t(a(o("common.formats.datetime",{value:Date.parse(s.connectTime??"")})),1)]),disconnected:e(({row:s})=>[s.disconnectTime?(r(),g(y,{key:0},[t(a(o("common.formats.datetime",{value:Date.parse(s.disconnectTime)})),1)],64)):_("",!0)]),responses:e(({row:s})=>{var l;return[(r(!0),g(y,null,B([((l=s.status)==null?void 0:l.total)??{}],p=>(r(),g(y,null,[t(a(p.responsesSent)+"/"+a(p.responsesAcknowledged),1)],64))),256))]}),_:2},1032,["headers","is-selected-row","items","onResize"]),n[13]||(n[13]=t()),d(f,null,{default:e(({Component:s})=>[k.child()?(r(),m(R,{key:0,width:"670px",onClose:function(){k.replace({name:"zone-ingress-detail-view",params:{zoneIngress:k.params.zoneIngress}})}},{default:e(()=>[(r(),m(D(s),{data:i.data.zoneIngressInsight.subscriptions},null,8,["data"]))]),_:2},1032,["onClose"])):_("",!0)]),_:2},1024)])):_("",!0)])]),_:2},1024)]),_:1})}}});export{F as default};

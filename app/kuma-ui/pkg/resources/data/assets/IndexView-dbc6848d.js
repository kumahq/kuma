import{d as S,r as m,o,i as a,w as t,j as l,p as R,n as d,E as $,H as p,a3 as C,l as f,F as I,k as b,$ as P,K as B,q as D,m as A,t as N}from"./index-8bdef5fd.js";import{A as T}from"./AppCollection-9c4ed2fa.js";import{S as E}from"./StatusBadge-e94c2294.js";import{S as K}from"./SummaryView-de117cc5.js";import{g as q}from"./dataplane-0a086c06.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-e4aed195.js";const F=S({__name:"IndexView",setup(L){function V(k){return k.map(u=>{const{name:_}=u,g={name:"zone-ingress-detail-view",params:{zoneIngress:_}},{networking:e}=u.zoneIngress;let z;e!=null&&e.address&&(e!=null&&e.port)&&(z=`${e.address}:${e.port}`);let c;e!=null&&e.advertisedAddress&&(e!=null&&e.advertisedPort)&&(c=`${e.advertisedAddress}:${e.advertisedPort}`);const y=q(u.zoneIngressInsight??{});return{detailViewRoute:g,name:_,addressPort:z,advertisedAddressPort:c,status:y}})}return(k,u)=>{const _=m("RouteTitle"),g=m("RouterLink"),e=m("KCard"),z=m("RouterView"),c=m("DataSource"),y=m("AppView"),h=m("RouteView");return o(),a(c,{src:"/me"},{default:t(({data:w})=>[w?(o(),a(h,{key:0,name:"zone-ingress-list-view",params:{zone:"",zoneIngress:""}},{default:t(({route:n,t:r})=>[l(y,null,{title:t(()=>[R("h2",null,[l(_,{title:r("zone-ingresses.routes.items.title"),render:!0},null,8,["title"])])]),default:t(()=>[d(),l(c,{src:`/zone-cps/${n.params.zone}/ingresses?page=1&size=100`},{default:t(({data:i,error:v})=>[l(e,null,{body:t(()=>[v!==void 0?(o(),a($,{key:0,error:v},null,8,["error"])):(o(),a(T,{key:1,class:"zone-ingress-collection","data-testid":"zone-ingress-collection",headers:[{label:"Name",key:"name"},{label:"Address",key:"addressPort"},{label:"Advertised address",key:"advertisedAddressPort"},{label:"Status",key:"status"},{label:"Details",key:"details",hideLabel:!0}],"page-number":1,"page-size":100,total:i==null?void 0:i.total,items:i?V(i.items):void 0,error:v,"empty-state-message":r("common.emptyState.message",{type:"Zone Ingresses"}),"empty-state-cta-to":r("zone-ingresses.href.docs"),"empty-state-cta-text":r("common.documentation"),"is-selected-row":s=>s.name===n.params.zoneIngress,onChange:n.update},{name:t(({row:s})=>[l(g,{to:{name:"zone-ingress-summary-view",params:{zone:n.params.zone,zoneIngress:s.name},query:{page:1,size:100}}},{default:t(()=>[d(p(s.name),1)]),_:2},1032,["to"])]),addressPort:t(({rowValue:s})=>[s?(o(),a(C,{key:0,text:s},null,8,["text"])):(o(),f(I,{key:1},[d(p(r("common.collection.none")),1)],64))]),advertisedAddressPort:t(({rowValue:s})=>[s?(o(),a(C,{key:0,text:s},null,8,["text"])):(o(),f(I,{key:1},[d(p(r("common.collection.none")),1)],64))]),status:t(({rowValue:s})=>[s?(o(),a(E,{key:0,status:s},null,8,["status"])):(o(),f(I,{key:1},[d(p(r("common.collection.none")),1)],64))]),details:t(({row:s})=>[l(g,{class:"details-link","data-testid":"details-link",to:{name:"zone-ingress-detail-view",params:{zoneIngress:s.name}}},{default:t(()=>[d(p(r("common.collection.details_link"))+" ",1),l(b(P),{display:"inline-block",decorative:"",size:b(B)},null,8,["size"])]),_:2},1032,["to"])]),_:2},1032,["total","items","error","empty-state-message","empty-state-cta-to","empty-state-cta-text","is-selected-row","onChange"]))]),_:2},1024),d(),n.params.zoneIngress?(o(),a(z,{key:0},{default:t(s=>[l(K,{onClose:x=>n.replace({name:"zone-ingress-list-view",params:{zone:n.params.zone},query:{page:1,size:100}})},{default:t(()=>[(o(),a(D(s.Component),{name:n.params.zoneIngress,"zone-ingress-overview":i==null?void 0:i.items.find(x=>x.name===n.params.zoneIngress)},null,8,["name","zone-ingress-overview"]))]),_:2},1032,["onClose"])]),_:2},1024)):A("",!0)]),_:2},1032,["src"])]),_:2},1024)]),_:1},8,["params"])):A("",!0)]),_:1})}}});const G=N(F,[["__scopeId","data-v-208a808b"]]);export{G as default};

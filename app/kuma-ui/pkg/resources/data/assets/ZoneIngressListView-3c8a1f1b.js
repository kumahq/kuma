import{K as b}from"./index-fce48c05.js";import{d as R,a as l,o,b as a,w as t,e as m,p as P,f as c,t as p,c as v,F as I,q as x,U as $,D as B,s as V,_ as D}from"./index-e9fbefd3.js";import{A as T}from"./AppCollection-98909f49.js";import{E as L}from"./ErrorBlock-a3710a04.js";import{S as N}from"./StatusBadge-494c559b.js";import{S as Z}from"./SummaryView-39651fdf.js";import{T as h}from"./TextWithCopyButton-0bfc7306.js";import{g as E}from"./dataplane-7a46b268.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-15e0e5b5.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-060e8475.js";import"./CopyButton-6f1494f2.js";const K=R({__name:"ZoneIngressListView",setup(q){function w(k){return k.map(_=>{const{name:u}=_,g={name:"zone-ingress-detail-view",params:{zoneIngress:u}},{networking:e}=_.zoneIngress;let z;e!=null&&e.address&&(e!=null&&e.port)&&(z=`${e.address}:${e.port}`);let d;e!=null&&e.advertisedAddress&&(e!=null&&e.advertisedPort)&&(d=`${e.advertisedAddress}:${e.advertisedPort}`);const f=E(_.zoneIngressInsight??{});return{detailViewRoute:g,name:u,addressPort:z,advertisedAddressPort:d,status:f}})}return(k,_)=>{const u=l("RouteTitle"),g=l("RouterLink"),e=l("KCard"),z=l("RouterView"),d=l("DataSource"),f=l("AppView"),A=l("RouteView");return o(),a(d,{src:"/me"},{default:t(({data:S})=>[S?(o(),a(A,{key:0,name:"zone-ingress-list-view",params:{zone:"",zoneIngress:""}},{default:t(({route:n,t:r})=>[m(f,null,{title:t(()=>[P("h2",null,[m(u,{title:r("zone-ingresses.routes.items.title")},null,8,["title"])])]),default:t(()=>[c(),m(d,{src:`/zone-cps/${n.params.zone}/ingresses?page=1&size=100`},{default:t(({data:i,error:y})=>[m(e,null,{default:t(()=>[y!==void 0?(o(),a(L,{key:0,error:y},null,8,["error"])):(o(),a(T,{key:1,class:"zone-ingress-collection","data-testid":"zone-ingress-collection",headers:[{label:"Name",key:"name"},{label:"Address",key:"addressPort"},{label:"Advertised address",key:"advertisedAddressPort"},{label:"Status",key:"status"},{label:"Details",key:"details",hideLabel:!0}],"page-number":1,"page-size":100,total:i==null?void 0:i.total,items:i?w(i.items):void 0,error:y,"empty-state-message":r("common.emptyState.message",{type:"Zone Ingresses"}),"empty-state-cta-to":r("zone-ingresses.href.docs"),"empty-state-cta-text":r("common.documentation"),"is-selected-row":s=>s.name===n.params.zoneIngress,onChange:n.update},{name:t(({row:s})=>[m(g,{to:{name:"zone-ingress-summary-view",params:{zone:n.params.zone,zoneIngress:s.name},query:{page:1,size:100}}},{default:t(()=>[c(p(s.name),1)]),_:2},1032,["to"])]),addressPort:t(({rowValue:s})=>[s?(o(),a(h,{key:0,text:s},null,8,["text"])):(o(),v(I,{key:1},[c(p(r("common.collection.none")),1)],64))]),advertisedAddressPort:t(({rowValue:s})=>[s?(o(),a(h,{key:0,text:s},null,8,["text"])):(o(),v(I,{key:1},[c(p(r("common.collection.none")),1)],64))]),status:t(({rowValue:s})=>[s?(o(),a(N,{key:0,status:s},null,8,["status"])):(o(),v(I,{key:1},[c(p(r("common.collection.none")),1)],64))]),details:t(({row:s})=>[m(g,{class:"details-link","data-testid":"details-link",to:{name:"zone-ingress-detail-view",params:{zoneIngress:s.name}}},{default:t(()=>[c(p(r("common.collection.details_link"))+" ",1),m(x($),{display:"inline-block",decorative:"",size:x(b)},null,8,["size"])]),_:2},1032,["to"])]),_:2},1032,["total","items","error","empty-state-message","empty-state-cta-to","empty-state-cta-text","is-selected-row","onChange"]))]),_:2},1024),c(),n.params.zoneIngress?(o(),a(z,{key:0},{default:t(s=>[m(Z,{onClose:C=>n.replace({name:"zone-ingress-list-view",params:{zone:n.params.zone},query:{page:1,size:100}})},{default:t(()=>[(o(),a(B(s.Component),{name:n.params.zoneIngress,"zone-ingress-overview":i==null?void 0:i.items.find(C=>C.name===n.params.zoneIngress)},null,8,["name","zone-ingress-overview"]))]),_:2},1032,["onClose"])]),_:2},1024)):V("",!0)]),_:2},1032,["src"])]),_:2},1024)]),_:1},8,["params"])):V("",!0)]),_:1})}}});const Y=D(K,[["__scopeId","data-v-01162f31"]]);export{Y as default};

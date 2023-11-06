import{d as I,r as i,o,i as r,w as s,j as m,p as b,n as c,E as R,H as z,a4 as B,l as v,F as w,k as C,a0 as A,K as D,q as N,m as E,t as T}from"./index-c6bd05ee.js";import{A as $}from"./AppCollection-6aabd095.js";import{S as K}from"./StatusBadge-07ca9e6a.js";import{S as q}from"./SummaryView-227f39d3.js";import{g as F}from"./dataplane-0a086c06.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-fb6e10fa.js";const L=I({__name:"IndexView",setup(P){function x(f){return f.map(p=>{const{name:d}=p,u={name:"zone-egress-detail-view",params:{zoneEgress:d}},{networking:t}=p.zoneEgress;let _;t!=null&&t.address&&(t!=null&&t.port)&&(_=`${t.address}:${t.port}`);const g=F(p.zoneEgressInsight??{});return{detailViewRoute:u,name:d,addressPort:_,status:g}})}return(f,p)=>{const d=i("RouteTitle"),u=i("RouterLink"),t=i("KCard"),_=i("RouterView"),g=i("DataSource"),V=i("AppView"),h=i("RouteView");return o(),r(g,{src:"/me"},{default:s(({data:S})=>[S?(o(),r(h,{key:0,name:"zone-egress-list-view",params:{zone:"",zoneEgress:""}},{default:s(({route:a,t:l})=>[m(V,null,{title:s(()=>[b("h2",null,[m(d,{title:l("zone-egresses.routes.items.title"),render:!0},null,8,["title"])])]),default:s(()=>[c(),m(g,{src:`/zone-cps/${a.params.zone||"*"}/egresses?page=1&size=100`},{default:s(({data:n,error:y})=>[m(t,null,{body:s(()=>[y!==void 0?(o(),r(R,{key:0,error:y},null,8,["error"])):(o(),r($,{key:1,class:"zone-egress-collection","data-testid":"zone-egress-collection",headers:[{label:"Name",key:"name"},{label:"Address",key:"addressPort"},{label:"Status",key:"status"},{label:"Details",key:"details",hideLabel:!0}],"page-number":1,"page-size":100,total:n==null?void 0:n.total,items:n?x(n.items):void 0,error:y,"empty-state-message":l("common.emptyState.message",{type:"Zone Egresses"}),"empty-state-cta-to":l("zone-egresses.href.docs"),"empty-state-cta-text":l("common.documentation"),"is-selected-row":e=>e.name===a.params.zoneEgress,onChange:a.update},{name:s(({row:e})=>[m(u,{to:{name:"zone-egress-summary-view",params:{zone:a.params.zone,zoneEgress:e.name},query:{page:1,size:100}}},{default:s(()=>[c(z(e.name),1)]),_:2},1032,["to"])]),addressPort:s(({rowValue:e})=>[e?(o(),r(B,{key:0,text:e},null,8,["text"])):(o(),v(w,{key:1},[c(z(l("common.collection.none")),1)],64))]),status:s(({rowValue:e})=>[e?(o(),r(K,{key:0,status:e},null,8,["status"])):(o(),v(w,{key:1},[c(z(l("common.collection.none")),1)],64))]),details:s(({row:e})=>[m(u,{class:"details-link","data-testid":"details-link",to:{name:"zone-egress-detail-view",params:{zoneEgress:e.name}}},{default:s(()=>[c(z(l("common.collection.details_link"))+" ",1),m(C(A),{display:"inline-block",decorative:"",size:C(D)},null,8,["size"])]),_:2},1032,["to"])]),_:2},1032,["total","items","error","empty-state-message","empty-state-cta-to","empty-state-cta-text","is-selected-row","onChange"]))]),_:2},1024),c(),a.params.zoneEgress?(o(),r(_,{key:0},{default:s(e=>[m(q,{onClose:k=>a.replace({name:"zone-egress-list-view",params:{zone:a.params.zone},query:{page:1,size:100}})},{default:s(()=>[(o(),r(N(e.Component),{name:a.params.zoneEgress,"zone-egress-overview":n==null?void 0:n.items.find(k=>k.name===a.params.zoneEgress)},null,8,["name","zone-egress-overview"]))]),_:2},1032,["onClose"])]),_:2},1024)):E("",!0)]),_:2},1032,["src"])]),_:2},1024)]),_:1},8,["params"])):E("",!0)]),_:1})}}});const G=T(L,[["__scopeId","data-v-d698c729"]]);export{G as default};

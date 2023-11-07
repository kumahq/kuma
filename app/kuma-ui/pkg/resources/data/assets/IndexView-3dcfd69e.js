import{d as b,a as i,o,b as r,w as s,e as m,p as I,f as c,E as R,t as z,a3 as B,c as v,F as w,q as C,$,K as D,G as N,v as E,_ as T}from"./index-70fb4e48.js";import{A}from"./AppCollection-08b32958.js";import{S as K}from"./StatusBadge-63893a0a.js";import{S as q}from"./SummaryView-e7da9a54.js";import{g as F}from"./dataplane-0a086c06.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-7812bd48.js";const L=b({__name:"IndexView",setup(P){function x(f){return f.map(p=>{const{name:d}=p,_={name:"zone-egress-detail-view",params:{zoneEgress:d}},{networking:t}=p.zoneEgress;let u;t!=null&&t.address&&(t!=null&&t.port)&&(u=`${t.address}:${t.port}`);const g=F(p.zoneEgressInsight??{});return{detailViewRoute:_,name:d,addressPort:u,status:g}})}return(f,p)=>{const d=i("RouteTitle"),_=i("RouterLink"),t=i("KCard"),u=i("RouterView"),g=i("DataSource"),V=i("AppView"),h=i("RouteView");return o(),r(g,{src:"/me"},{default:s(({data:S})=>[S?(o(),r(h,{key:0,name:"zone-egress-list-view",params:{zone:"",zoneEgress:""}},{default:s(({route:a,t:l})=>[m(V,null,{title:s(()=>[I("h2",null,[m(d,{title:l("zone-egresses.routes.items.title"),render:!0},null,8,["title"])])]),default:s(()=>[c(),m(g,{src:`/zone-cps/${a.params.zone||"*"}/egresses?page=1&size=100`},{default:s(({data:n,error:y})=>[m(t,null,{body:s(()=>[y!==void 0?(o(),r(R,{key:0,error:y},null,8,["error"])):(o(),r(A,{key:1,class:"zone-egress-collection","data-testid":"zone-egress-collection",headers:[{label:"Name",key:"name"},{label:"Address",key:"addressPort"},{label:"Status",key:"status"},{label:"Details",key:"details",hideLabel:!0}],"page-number":1,"page-size":100,total:n==null?void 0:n.total,items:n?x(n.items):void 0,error:y,"empty-state-message":l("common.emptyState.message",{type:"Zone Egresses"}),"empty-state-cta-to":l("zone-egresses.href.docs"),"empty-state-cta-text":l("common.documentation"),"is-selected-row":e=>e.name===a.params.zoneEgress,onChange:a.update},{name:s(({row:e})=>[m(_,{to:{name:"zone-egress-summary-view",params:{zone:a.params.zone,zoneEgress:e.name},query:{page:1,size:100}}},{default:s(()=>[c(z(e.name),1)]),_:2},1032,["to"])]),addressPort:s(({rowValue:e})=>[e?(o(),r(B,{key:0,text:e},null,8,["text"])):(o(),v(w,{key:1},[c(z(l("common.collection.none")),1)],64))]),status:s(({rowValue:e})=>[e?(o(),r(K,{key:0,status:e},null,8,["status"])):(o(),v(w,{key:1},[c(z(l("common.collection.none")),1)],64))]),details:s(({row:e})=>[m(_,{class:"details-link","data-testid":"details-link",to:{name:"zone-egress-detail-view",params:{zoneEgress:e.name}}},{default:s(()=>[c(z(l("common.collection.details_link"))+" ",1),m(C($),{display:"inline-block",decorative:"",size:C(D)},null,8,["size"])]),_:2},1032,["to"])]),_:2},1032,["total","items","error","empty-state-message","empty-state-cta-to","empty-state-cta-text","is-selected-row","onChange"]))]),_:2},1024),c(),a.params.zoneEgress?(o(),r(u,{key:0},{default:s(e=>[m(q,{onClose:k=>a.replace({name:"zone-egress-list-view",params:{zone:a.params.zone},query:{page:1,size:100}})},{default:s(()=>[(o(),r(N(e.Component),{name:a.params.zoneEgress,"zone-egress-overview":n==null?void 0:n.items.find(k=>k.name===a.params.zoneEgress)},null,8,["name","zone-egress-overview"]))]),_:2},1032,["onClose"])]),_:2},1024)):E("",!0)]),_:2},1032,["src"])]),_:2},1024)]),_:1},8,["params"])):E("",!0)]),_:1})}}});const H=T(L,[["__scopeId","data-v-d698c729"]]);export{H as default};

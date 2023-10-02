import{d as x,r as s,o as n,g as r,w as e,h as a,m as S,l as _,E as T,Y as B,C as f,a1 as D,j as v,F as E,S as L,i as C,Z as N,K as A,k as Z,q as O}from"./index-65a641bf.js";import{g as $}from"./dataplane-a974028d.js";const F=x({__name:"ZoneEgressListView",setup(P){function I(z){return z.map(m=>{const{name:p}=m,u={name:"zone-egress-detail-view",params:{zoneEgress:p}},{networking:t}=m.zoneEgress;let d;t!=null&&t.address&&(t!=null&&t.port)&&(d=`${t.address}:${t.port}`);const g=$(m.zoneEgressInsight??{});return{detailViewRoute:u,name:p,addressPort:d,status:g}})}return(z,m)=>{const p=s("RouteTitle"),u=s("RouterLink"),t=s("KIcon"),d=s("KButton"),g=s("KDropdownItem"),b=s("KDropdownMenu"),h=s("KCard"),w=s("DataSource"),K=s("AppView"),R=s("RouteView");return n(),r(w,{src:"/me"},{default:e(({data:k})=>[k?(n(),r(R,{key:0,name:"zone-egress-list-view",params:{page:1,size:k.pageSize}},{default:e(({route:i,t:c})=>[a(K,null,{title:e(()=>[S("h1",null,[a(p,{title:c("zone-egresses.routes.items.title"),render:!0},null,8,["title"])])]),default:e(()=>[_(),a(w,{src:`/zone-egress-overviews?page=${i.params.page}&size=${i.params.size}`},{default:e(({data:l,error:y})=>[a(h,null,{body:e(()=>[y!==void 0?(n(),r(T,{key:0,error:y},null,8,["error"])):(n(),r(B,{key:1,class:"zone-egress-collection","data-testid":"zone-egress-collection",headers:[{label:"Name",key:"name"},{label:"Address",key:"addressPort"},{label:"Status",key:"status"},{label:"Actions",key:"actions",hideLabel:!0}],"page-number":parseInt(i.params.page),"page-size":parseInt(i.params.size),total:l==null?void 0:l.total,items:l?I(l.items):void 0,error:y,"empty-state-message":c("common.emptyState.message",{type:"Zone Egresses"}),"empty-state-cta-to":c("zone-egresses.href.docs"),"empty-state-cta-text":c("common.documentation"),onChange:i.update},{name:e(({row:o,rowValue:V})=>[a(u,{to:o.detailViewRoute,"data-testid":"detail-view-link"},{default:e(()=>[_(f(V),1)]),_:2},1032,["to"])]),addressPort:e(({rowValue:o})=>[o?(n(),r(D,{key:0,text:o},null,8,["text"])):(n(),v(E,{key:1},[_(f(c("common.collection.none")),1)],64))]),status:e(({rowValue:o})=>[o?(n(),r(L,{key:0,status:o},null,8,["status"])):(n(),v(E,{key:1},[_(f(c("common.collection.none")),1)],64))]),actions:e(({row:o})=>[a(b,{class:"actions-dropdown","data-testid":"actions-dropdown","kpop-attributes":{placement:"bottomEnd",popoverClasses:"mt-5 more-actions-popover"},width:"150"},{default:e(()=>[a(d,{class:"non-visual-button",appearance:"secondary",size:"small"},{icon:e(()=>[a(t,{color:C(N),icon:"more",size:C(A)},null,8,["color","size"])]),_:1})]),items:e(()=>[a(g,{item:{to:o.detailViewRoute,label:c("common.collection.actions.view")}},null,8,["item"])]),_:2},1024)]),_:2},1032,["page-number","page-size","total","items","error","empty-state-message","empty-state-cta-to","empty-state-cta-text","onChange"]))]),_:2},1024)]),_:2},1032,["src"])]),_:2},1024)]),_:2},1032,["params"])):Z("",!0)]),_:1})}}});const j=O(F,[["__scopeId","data-v-cf9f5951"]]);export{j as default};

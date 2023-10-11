import{d as B,r as s,o as n,g as l,w as e,h as a,m as D,l as d,E,Z as R,D as y,a4 as S,j as z,F as x,U as N,i as h,a0 as T,K as $,k as A,q as F}from"./index-213666ad.js";import{g as L}from"./dataplane-0a086c06.js";const P=B({__name:"IndexView",setup(Z){function v(f){return f.map(i=>{const{name:m}=i,u={name:"zone-egress-detail-view",params:{zoneEgress:m}},{networking:t}=i.zoneEgress;let p;t!=null&&t.address&&(t!=null&&t.port)&&(p=`${t.address}:${t.port}`);const _=L(i.zoneEgressInsight??{});return{detailViewRoute:u,name:m,addressPort:p,status:_}})}return(f,i)=>{const m=s("RouteTitle"),u=s("RouterLink"),t=s("KButton"),p=s("KDropdownItem"),_=s("KDropdownMenu"),b=s("KCard"),w=s("DataSource"),C=s("AppView"),I=s("RouteView");return n(),l(w,{src:"/me"},{default:e(({data:V})=>[V?(n(),l(I,{key:0,name:"zone-egress-list-view",params:{zone:""}},{default:e(({route:k,t:c})=>[a(C,null,{title:e(()=>[D("h2",null,[a(m,{title:c("zone-egresses.routes.items.title"),render:!0},null,8,["title"])])]),default:e(()=>[d(),a(w,{src:`/zone-cps/${k.params.zone||"*"}/egresses?page=1&size=100`},{default:e(({data:r,error:g})=>[a(b,null,{body:e(()=>[g!==void 0?(n(),l(E,{key:0,error:g},null,8,["error"])):(n(),l(R,{key:1,class:"zone-egress-collection","data-testid":"zone-egress-collection",headers:[{label:"Name",key:"name"},{label:"Address",key:"addressPort"},{label:"Status",key:"status"},{label:"Actions",key:"actions",hideLabel:!0}],"page-number":1,"page-size":100,total:r==null?void 0:r.total,items:r?v(r.items):void 0,error:g,"empty-state-message":c("common.emptyState.message",{type:"Zone Egresses"}),"empty-state-cta-to":c("zone-egresses.href.docs"),"empty-state-cta-text":c("common.documentation"),onChange:k.update},{name:e(({row:o,rowValue:K})=>[a(u,{to:o.detailViewRoute,"data-testid":"detail-view-link"},{default:e(()=>[d(y(K),1)]),_:2},1032,["to"])]),addressPort:e(({rowValue:o})=>[o?(n(),l(S,{key:0,text:o},null,8,["text"])):(n(),z(x,{key:1},[d(y(c("common.collection.none")),1)],64))]),status:e(({rowValue:o})=>[o?(n(),l(N,{key:0,status:o},null,8,["status"])):(n(),z(x,{key:1},[d(y(c("common.collection.none")),1)],64))]),actions:e(({row:o})=>[a(_,{class:"actions-dropdown","data-testid":"actions-dropdown","kpop-attributes":{placement:"bottomEnd",popoverClasses:"mt-5 more-actions-popover"},width:"150"},{default:e(()=>[a(t,{class:"non-visual-button",appearance:"secondary",size:"small"},{default:e(()=>[a(h(T),{size:h($)},null,8,["size"])]),_:1})]),items:e(()=>[a(p,{item:{to:o.detailViewRoute,label:c("common.collection.actions.view")}},null,8,["item"])]),_:2},1024)]),_:2},1032,["total","items","error","empty-state-message","empty-state-cta-to","empty-state-cta-text","onChange"]))]),_:2},1024)]),_:2},1032,["src"])]),_:2},1024)]),_:1},8,["params"])):A("",!0)]),_:1})}}});const j=F(P,[["__scopeId","data-v-d05c1119"]]);export{j as default};

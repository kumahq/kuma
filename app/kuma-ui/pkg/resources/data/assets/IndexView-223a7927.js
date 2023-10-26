import{d as B,r as s,o as n,i as c,w as e,j as a,p as E,n as d,E as R,H as y,a1 as S,l as z,F as x,k as v,$ as D,K as A,m as N,t as T}from"./index-0d828147.js";import{A as $}from"./AppCollection-640ff5f7.js";import{S as F}from"./StatusBadge-e02331a5.js";import{g as L}from"./dataplane-0a086c06.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-25f4cd1f.js";const P=B({__name:"IndexView",setup(M){function b(f){return f.map(m=>{const{name:i}=m,u={name:"zone-egress-detail-view",params:{zoneEgress:i}},{networking:t}=m.zoneEgress;let p;t!=null&&t.address&&(t!=null&&t.port)&&(p=`${t.address}:${t.port}`);const _=L(m.zoneEgressInsight??{});return{detailViewRoute:u,name:i,addressPort:p,status:_}})}return(f,m)=>{const i=s("RouteTitle"),u=s("RouterLink"),t=s("KButton"),p=s("KDropdownItem"),_=s("KDropdownMenu"),h=s("KCard"),w=s("DataSource"),C=s("AppView"),I=s("RouteView");return n(),c(w,{src:"/me"},{default:e(({data:V})=>[V?(n(),c(I,{key:0,name:"zone-egress-list-view",params:{zone:""}},{default:e(({route:k,t:r})=>[a(C,null,{title:e(()=>[E("h2",null,[a(i,{title:r("zone-egresses.routes.items.title"),render:!0},null,8,["title"])])]),default:e(()=>[d(),a(w,{src:`/zone-cps/${k.params.zone||"*"}/egresses?page=1&size=100`},{default:e(({data:l,error:g})=>[a(h,null,{body:e(()=>[g!==void 0?(n(),c(R,{key:0,error:g},null,8,["error"])):(n(),c($,{key:1,class:"zone-egress-collection","data-testid":"zone-egress-collection",headers:[{label:"Name",key:"name"},{label:"Address",key:"addressPort"},{label:"Status",key:"status"},{label:"Actions",key:"actions",hideLabel:!0}],"page-number":1,"page-size":100,total:l==null?void 0:l.total,items:l?b(l.items):void 0,error:g,"empty-state-message":r("common.emptyState.message",{type:"Zone Egresses"}),"empty-state-cta-to":r("zone-egresses.href.docs"),"empty-state-cta-text":r("common.documentation"),onChange:k.update},{name:e(({row:o,rowValue:K})=>[a(u,{to:o.detailViewRoute,"data-testid":"detail-view-link"},{default:e(()=>[d(y(K),1)]),_:2},1032,["to"])]),addressPort:e(({rowValue:o})=>[o?(n(),c(S,{key:0,text:o},null,8,["text"])):(n(),z(x,{key:1},[d(y(r("common.collection.none")),1)],64))]),status:e(({rowValue:o})=>[o?(n(),c(F,{key:0,status:o},null,8,["status"])):(n(),z(x,{key:1},[d(y(r("common.collection.none")),1)],64))]),actions:e(({row:o})=>[a(_,{class:"actions-dropdown","data-testid":"actions-dropdown","kpop-attributes":{placement:"bottomEnd",popoverClasses:"mt-5 more-actions-popover"},width:"150"},{default:e(()=>[a(t,{class:"non-visual-button",appearance:"secondary",size:"small"},{default:e(()=>[a(v(D),{size:v(A)},null,8,["size"])]),_:1})]),items:e(()=>[a(p,{item:{to:o.detailViewRoute,label:r("common.collection.actions.view")}},null,8,["item"])]),_:2},1024)]),_:2},1032,["total","items","error","empty-state-message","empty-state-cta-to","empty-state-cta-text","onChange"]))]),_:2},1024)]),_:2},1032,["src"])]),_:2},1024)]),_:1},8,["params"])):N("",!0)]),_:1})}}});const U=T(P,[["__scopeId","data-v-7274859c"]]);export{U as default};

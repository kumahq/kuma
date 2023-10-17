import{d as T,r as s,o as n,i as c,w as t,j as a,p as B,n as l,E as S,a0 as D,H as u,a5 as b,l as g,F as I,W as E,k as w,aJ as N,K as P,m as $,t as L}from"./index-78eccadf.js";import{g as O}from"./dataplane-0a086c06.js";const F=T({__name:"IndexView",setup(U){function C(z){return z.map(i=>{const{name:m}=i,y={name:"zone-ingress-detail-view",params:{zoneIngress:m}},{networking:e}=i.zoneIngress;let p;e!=null&&e.address&&(e!=null&&e.port)&&(p=`${e.address}:${e.port}`);let _;e!=null&&e.advertisedAddress&&(e!=null&&e.advertisedPort)&&(_=`${e.advertisedAddress}:${e.advertisedPort}`);const f=O(i.zoneIngressInsight??{});return{detailViewRoute:y,name:m,addressPort:p,advertisedAddressPort:_,status:f}})}return(z,i)=>{const m=s("RouteTitle"),y=s("RouterLink"),e=s("KIcon"),p=s("KButton"),_=s("KDropdownItem"),f=s("KDropdownMenu"),K=s("KCard"),k=s("DataSource"),R=s("AppView"),h=s("RouteView");return n(),c(k,{src:"/me"},{default:t(({data:A})=>[A?(n(),c(h,{key:0,name:"zone-ingress-list-view",params:{zone:""}},{default:t(({route:x,t:r})=>[a(R,null,{title:t(()=>[B("h2",null,[a(m,{title:r("zone-ingresses.routes.items.title"),render:!0},null,8,["title"])])]),default:t(()=>[l(),a(k,{src:`/zone-cps/${x.params.zone}/ingresses?page=1&size=100`},{default:t(({data:d,error:v})=>[a(K,null,{body:t(()=>[v!==void 0?(n(),c(S,{key:0,error:v},null,8,["error"])):(n(),c(D,{key:1,class:"zone-ingress-collection","data-testid":"zone-ingress-collection",headers:[{label:"Name",key:"name"},{label:"Address",key:"addressPort"},{label:"Advertised address",key:"advertisedAddressPort"},{label:"Status",key:"status"},{label:"Actions",key:"actions",hideLabel:!0}],"page-number":1,"page-size":100,total:d==null?void 0:d.total,items:d?C(d.items):void 0,error:v,"empty-state-message":r("common.emptyState.message",{type:"Zone Ingresses"}),"empty-state-cta-to":r("zone-ingresses.href.docs"),"empty-state-cta-text":r("common.documentation"),onChange:x.update},{name:t(({row:o,rowValue:V})=>[a(y,{to:o.detailViewRoute,"data-testid":"detail-view-link"},{default:t(()=>[l(u(V),1)]),_:2},1032,["to"])]),addressPort:t(({rowValue:o})=>[o?(n(),c(b,{key:0,text:o},null,8,["text"])):(n(),g(I,{key:1},[l(u(r("common.collection.none")),1)],64))]),advertisedAddressPort:t(({rowValue:o})=>[o?(n(),c(b,{key:0,text:o},null,8,["text"])):(n(),g(I,{key:1},[l(u(r("common.collection.none")),1)],64))]),status:t(({rowValue:o})=>[o?(n(),c(E,{key:0,status:o},null,8,["status"])):(n(),g(I,{key:1},[l(u(r("common.collection.none")),1)],64))]),actions:t(({row:o})=>[a(f,{class:"actions-dropdown","data-testid":"actions-dropdown","kpop-attributes":{placement:"bottomEnd",popoverClasses:"mt-5 more-actions-popover"},width:"150"},{default:t(()=>[a(p,{class:"non-visual-button",appearance:"secondary",size:"small"},{icon:t(()=>[a(e,{color:w(N),icon:"more",size:w(P)},null,8,["color","size"])]),_:1})]),items:t(()=>[a(_,{item:{to:o.detailViewRoute,label:r("common.collection.actions.view")}},null,8,["item"])]),_:2},1024)]),_:2},1032,["total","items","error","empty-state-message","empty-state-cta-to","empty-state-cta-text","onChange"]))]),_:2},1024)]),_:2},1032,["src"])]),_:2},1024)]),_:1},8,["params"])):$("",!0)]),_:1})}}});const Z=L(F,[["__scopeId","data-v-f0fbdc79"]]);export{Z as default};

import{d as L,R as O,y as Z,L as B,h as t,o as _,a as d,w as e,j as n,z as a,V as w,g as f,X as S,k as i,B as K,t as u,r as $,Y as M,e as z,W as C}from"./index-ChMk9xbI.js";import{_ as P}from"./DeleteResourceModal.vue_vue_type_script_setup_true_lang-rIrZi3gT.js";const W=L({__name:"ZoneDetailTabsView",setup(Y){const h=O(),{t:o}=Z(),b=B([]),g=c=>{const r=[];c.zoneInsight.store==="memory"&&r.push({kind:"ZONE_STORE_TYPE_MEMORY",payload:{}}),C(c.zoneInsight,"version.kumaCp.kumaCpGlobalCompatible","true")||r.push({kind:"INCOMPATIBLE_ZONE_AND_GLOBAL_CPS_VERSIONS",payload:{zoneCpVersion:C(c.zoneInsight,"version.kumaCp.version",o("common.collection.none"))}}),b.value=r};async function k(c){await h.deleteZone({name:c})}return(c,r)=>{const x=t("RouteTitle"),v=t("KDropdownItem"),D=t("XTeleportTemplate"),V=t("XDisclosure"),y=t("KDropdown"),T=t("XAction"),R=t("XTabs"),A=t("RouterView"),I=t("AppView"),N=t("DataLoader"),X=t("RouteView");return _(),d(X,{name:"zone-cp-detail-tabs-view",params:{zone:""}},{default:e(({can:E,route:l})=>[n(N,{src:`/zone-cps/${l.params.zone}`,onChange:g},{default:e(({data:m})=>[m?(_(),d(I,{key:0,breadcrumbs:[{to:{name:"zone-cp-list-view"},text:a(o)("zone-cps.routes.item.breadcrumbs")}]},w({title:e(()=>[f("h1",null,[n(S,{text:l.params.zone},{default:e(()=>[n(x,{title:a(o)("zone-cps.routes.item.title",{name:l.params.zone})},null,8,["title"])]),_:2},1032,["text"])])]),default:e(()=>{var p;return[i(),i(),n(R,{selected:(p=l.child())==null?void 0:p.name},w({_:2},[K(l.children,({name:s})=>({name:`${s}-tab`,fn:e(()=>[n(T,{to:{name:s}},{default:e(()=>[i(u(a(o)(`zone-cps.routes.item.navigation.${s}`)),1)]),_:2},1032,["to"])])}))]),1032,["selected"]),i(),n(A,null,{default:e(s=>[(_(),d($(s.Component),{data:m,notifications:b.value},null,8,["data","notifications"]))]),_:2},1024)]}),_:2},[E("create zones")?{name:"actions",fn:e(()=>[n(y,{"kpop-attributes":{placement:"bottomEnd"},"trigger-text":a(o)("zones.action_menu.toggle_button"),"show-caret":"",width:"280"},{items:e(()=>[n(V,null,{default:e(({expanded:p,toggle:s})=>[n(v,{danger:"","data-testid":"delete-button",onClick:M(s,["prevent"])},{default:e(()=>[i(u(a(o)("zones.action_menu.delete_button")),1)]),_:2},1032,["onClick"]),i(),n(D,{to:{name:"modal-layer"}},{default:e(()=>[p?(_(),d(P,{key:0,"confirmation-text":m.name,"delete-function":()=>k(m.name),"is-visible":"","action-button-text":a(o)("common.delete_modal.proceed_button"),title:a(o)("common.delete_modal.title",{type:"Zone"}),"data-testid":"delete-zone-modal",onCancel:s,onDelete:()=>l.replace({name:"zone-cp-list-view"})},{default:e(()=>[f("p",null,u(a(o)("common.delete_modal.text1",{type:"Zone",name:m.name})),1),i(),f("p",null,u(a(o)("common.delete_modal.text2")),1)]),_:2},1032,["confirmation-text","delete-function","action-button-text","title","onCancel","onDelete"])):z("",!0)]),_:2},1024)]),_:2},1024)]),_:2},1032,["trigger-text"])]),key:"0"}:void 0]),1032,["breadcrumbs"])):z("",!0)]),_:2},1032,["src"])]),_:1})}}});export{W as default};

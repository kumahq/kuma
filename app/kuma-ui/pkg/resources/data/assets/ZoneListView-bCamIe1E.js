import{d as U,R as W,D as R,a as c,o,b as g,w as e,e as t,m as I,f as a,X as H,c as _,F as f,p as b,S as J,t as r,q as C,O as M,$ as B,G as V,N as Q,K as Y,a0 as ee,E as ne,_ as te}from"./index-UmH8j8ci.js";import{A as oe}from"./AppCollection-oxbvtryr.js";import{_ as se}from"./DeleteResourceModal.vue_vue_type_script_setup_true_lang-B2PhRXVF.js";import{S as ae}from"./StatusBadge-Cpcdk_fe.js";import{S as le}from"./SummaryView-DX0YbGlV.js";const ie=["data-testid"],re=U({__name:"ZoneListView",setup(ce){const E=W(),x=R({}),K=R({}),L=k=>{const w="zoneIngress";x.value=k.items.reduce((d,m)=>{var z;const p=(z=m[w])==null?void 0:z.zone;if(typeof p<"u"){typeof d[p]>"u"&&(d[p]={online:[],offline:[]});const v=typeof m[`${w}Insight`].connectedSubscription<"u"?"online":"offline";d[p][v].push(m)}return d},{})},T=k=>{const w="zoneEgress";K.value=k.items.reduce((d,m)=>{var z;const p=(z=m[w])==null?void 0:z.zone;if(typeof p<"u"){typeof d[p]>"u"&&(d[p]={online:[],offline:[]});const v=typeof m[`${w}Insight`].connectedSubscription<"u"?"online":"offline";d[p][v].push(m)}return d},{})};async function Z(k){await E.deleteZone({name:k})}return(k,w)=>{const d=c("RouteTitle"),m=c("DataSource"),p=c("KButton"),z=c("XTeleportTemplate"),v=c("RouterLink"),A=c("XIcon"),N=c("KDropdownItem"),X=c("XDisclosure"),q=c("KDropdown"),O=c("KCard"),F=c("RouterView"),G=c("AppView"),j=c("RouteView");return o(),g(m,{src:"/me"},{default:e(({data:$})=>[$?(o(),g(j,{key:0,name:"zone-cp-list-view",params:{page:1,size:$.pageSize,zone:""}},{default:e(({route:l,t:i,can:h})=>[t(G,null,{title:e(()=>[I("h1",null,[t(d,{title:i("zone-cps.routes.items.title")},null,8,["title"])])]),default:e(()=>[a(),t(m,{src:`/zone-cps?page=${l.params.page}&size=${l.params.size}`},{default:e(({data:u,error:S,refresh:P})=>[t(m,{src:"/zone-ingress-overviews?page=1&size=100",onChange:L}),a(),t(m,{src:"/zone-egress-overviews?page=1&size=100",onChange:T}),a(),t(O,null,{default:e(()=>[S!==void 0?(o(),g(H,{key:0,error:S},null,8,["error"])):(o(),_(f,{key:1},[h("create zones")&&((u==null?void 0:u.items)??[]).length>0?(o(),g(z,{key:0,to:{name:"zone-cp-list-view-actions"}},{default:e(()=>[t(p,{appearance:"primary",to:{name:"zone-create-view"},"data-testid":"create-zone-link"},{default:e(()=>[t(b(J)),a(" "+r(i("zones.index.create")),1)]),_:2},1024)]),_:2},1024)):C("",!0),a(),t(oe,{class:"zone-cp-collection","data-testid":"zone-cp-collection",headers:[{label:"Name",key:"name"},{label:"Zone CP Version",key:"zoneCpVersion"},{label:"Type",key:"type"},{label:"Ingresses (online / total)",key:"ingress"},{label:"Egresses (online / total)",key:"egress"},{label:"Status",key:"state"},{label:"Warnings",key:"warnings",hideLabel:!0},{label:"Details",key:"details",hideLabel:!0},{label:"Actions",key:"actions",hideLabel:!0}],"page-number":l.params.page,"page-size":l.params.size,total:u==null?void 0:u.total,items:u==null?void 0:u.items,error:S,"empty-state-title":h("create zones")?i("zone-cps.empty_state.title"):void 0,"empty-state-message":h("create zones")?i("zone-cps.empty_state.message"):void 0,"empty-state-cta-to":h("create zones")?{name:"zone-create-view"}:void 0,"empty-state-cta-text":h("create zones")?i("zones.index.create"):void 0,"is-selected-row":n=>n.name===l.params.zone,onChange:l.update},M({name:e(({row:n})=>[t(v,{to:{name:"zone-cp-detail-view",params:{zone:n.name},query:{page:l.params.page,size:l.params.size}}},{default:e(()=>[a(r(n.name),1)]),_:2},1032,["to"])]),zoneCpVersion:e(({row:n})=>[a(r(b(B)(n.zoneInsight,"version.kumaCp.version",i("common.collection.none"))),1)]),type:e(({row:n})=>[a(r(n.zoneInsight.environment.length>0?n.zoneInsight.environment:"kubernetes"),1)]),ingress:e(({row:n})=>[(o(!0),_(f,null,V([x.value[n.name]||{online:[],offline:[]}],s=>(o(),_(f,null,[a(r(s.online.length)+" / "+r(s.online.length+s.offline.length),1)],64))),256))]),egress:e(({row:n})=>[(o(!0),_(f,null,V([K.value[n.name]||{online:[],offline:[]}],s=>(o(),_(f,null,[a(r(s.online.length)+" / "+r(s.online.length+s.offline.length),1)],64))),256))]),state:e(({row:n})=>[t(ae,{status:n.state},null,8,["status"])]),warnings:e(({row:n})=>[(o(!0),_(f,null,V([{version_mismatch:!b(B)(n.zoneInsight,"version.kumaCp.kumaCpGlobalCompatible","true"),store_memory:n.zoneInsight.store==="memory"}],s=>(o(),_(f,{key:`${s.version_mismatch}-${s.store_memory}`},[Object.values(s).some(y=>y)?(o(),g(A,{key:0,name:"warning","data-testid":"warning"},{default:e(()=>[I("ul",null,[(o(!0),_(f,null,V(s,(y,D)=>(o(),_(f,{key:D},[y?(o(),_("li",{key:0,"data-testid":`warning-${D}`},r(i(`zone-cps.list.${D}`)),9,ie)):C("",!0)],64))),128))])]),_:2},1024)):(o(),_(f,{key:1},[a(r(i("common.collection.none")),1)],64))],64))),128))]),details:e(({row:n})=>[t(v,{class:"details-link","data-testid":"details-link",to:{name:"zone-cp-detail-view",params:{zone:n.name}}},{default:e(()=>[a(r(i("common.collection.details_link"))+" ",1),t(b(Q),{decorative:"",size:b(Y)},null,8,["size"])]),_:2},1032,["to"])]),_:2},[h("create zones")?{name:"actions",fn:e(({row:n})=>[t(q,{class:"actions-dropdown","kpop-attributes":{placement:"bottomEnd",popoverClasses:"mt-5 more-actions-popover"},width:"150"},{default:e(()=>[t(p,{class:"non-visual-button",appearance:"secondary","icon-only":""},{default:e(()=>[t(b(ee))]),_:1})]),items:e(()=>[t(X,null,{default:e(({expanded:s,toggle:y})=>[t(N,{danger:"","data-testid":"dropdown-delete-item",onClick:y},{default:e(()=>[a(r(i("common.collection.actions.delete")),1)]),_:2},1032,["onClick"]),a(),t(z,{to:{name:"modal-layer"}},{default:e(()=>[s?(o(),g(se,{key:0,"confirmation-text":n.name,"delete-function":()=>Z(n.name),"is-visible":"","action-button-text":i("common.delete_modal.proceed_button"),title:i("common.delete_modal.title",{type:"Zone"}),"data-testid":"delete-zone-modal",onCancel:y,onDelete:()=>{y(),P()}},{default:e(()=>[I("p",null,r(i("common.delete_modal.text1",{type:"Zone",name:n.name})),1),a(),I("p",null,r(i("common.delete_modal.text2")),1)]),_:2},1032,["confirmation-text","delete-function","action-button-text","title","onCancel","onDelete"])):C("",!0)]),_:2},1024)]),_:2},1024)]),_:2},1024)]),key:"0"}:void 0]),1032,["page-number","page-size","total","items","error","empty-state-title","empty-state-message","empty-state-cta-to","empty-state-cta-text","is-selected-row","onChange"])],64))]),_:2},1024),a(),l.params.zone?(o(),g(F,{key:0},{default:e(n=>[t(le,{onClose:s=>l.replace({name:"zone-cp-list-view",query:{page:l.params.page,size:l.params.size}})},{default:e(()=>[(o(),g(ne(n.Component),{name:l.params.zone,"zone-overview":u==null?void 0:u.items.find(s=>s.name===l.params.zone)},null,8,["name","zone-overview"]))]),_:2},1032,["onClose"])]),_:2},1024)):C("",!0)]),_:2},1032,["src"])]),_:2},1024)]),_:2},1032,["params"])):C("",!0)]),_:1})}}}),fe=te(re,[["__scopeId","data-v-36a10291"]]);export{fe as default};

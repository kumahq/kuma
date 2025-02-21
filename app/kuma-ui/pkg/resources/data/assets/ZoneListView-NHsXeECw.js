import{d as K,x as A,r as i,o as r,q as _,w as e,b as s,m as L,e as a,p as R,F as O,t as p,s as b,B as Q,c as y,M as w,N as v,W as U,S as Y,I as ee,_ as ne}from"./index-DxWkv34s.js";import{S as oe}from"./SummaryView-DD4ymD2U.js";const te=["data-testid"],se=K({__name:"ZoneListView",setup(ae){const V=A({}),X=A({}),N=C=>{const n="zoneIngress";V.value=C.items.reduce((u,d)=>{var f;const m=(f=d[n])==null?void 0:f.zone;if(typeof m<"u"){typeof u[m]>"u"&&(u[m]={online:[],offline:[]});const k=typeof d[`${n}Insight`].connectedSubscription<"u"?"online":"offline";u[m][k].push(d)}return u},{})},T=C=>{const n="zoneEgress";X.value=C.items.reduce((u,d)=>{var f;const m=(f=d[n])==null?void 0:f.zone;if(typeof m<"u"){typeof u[m]>"u"&&(u[m]={online:[],offline:[]});const k=typeof d[`${n}Insight`].connectedSubscription<"u"?"online":"offline";u[m][k].push(d)}return u},{})};return(C,n)=>{const u=i("RouteTitle"),d=i("DataSource"),m=i("XI18n"),f=i("XAction"),k=i("XTeleportTemplate"),S=i("XIcon"),I=i("DataLoader"),$=i("XPrompt"),x=i("DataSink"),B=i("XDisclosure"),Z=i("XActionGroup"),E=i("DataCollection"),q=i("XCard"),P=i("RouterView"),F=i("AppView"),G=i("RouteView");return r(),_(G,{name:"zone-cp-list-view",params:{page:1,size:Number,zone:""}},{default:e(({route:c,t:g,can:D,uri:W,me:z})=>[s(F,{docs:g("zones.href.docs.cta")},{title:e(()=>[L("h1",null,[s(u,{title:g("zone-cps.routes.items.title")},null,8,["title"])])]),default:e(()=>[n[16]||(n[16]=a()),s(d,{src:W(R(O),"/zone-cps",{},{page:c.params.page,size:c.params.size})},{default:e(({data:l,error:M,refresh:j})=>[s(d,{src:"/zone-ingress-overviews?page=1&size=100",onChange:N}),n[12]||(n[12]=a()),s(d,{src:"/zone-egress-overviews?page=1&size=100",onChange:T}),n[13]||(n[13]=a()),s(m,{path:"zone-cps.routes.items.intro","default-path":"common.i18n.ignore-error"}),n[14]||(n[14]=a()),s(q,null,{default:e(()=>[D("create zones")&&((l==null?void 0:l.items)??[]).length>0?(r(),_(k,{key:0,to:{name:"zone-cp-list-view-actions"}},{default:e(()=>[s(f,{action:"create",appearance:"primary",to:{name:"zone-create-view"},"data-testid":"create-zone-link"},{default:e(()=>[a(p(g("zones.index.create")),1)]),_:2},1024)]),_:2},1024)):b("",!0),n[11]||(n[11]=a()),s(I,{data:[l],errors:[M]},{loadable:e(()=>[s(E,{type:"zone-cps",items:(l==null?void 0:l.items)??[void 0],page:c.params.page,"page-size":c.params.size,total:l==null?void 0:l.total,onChange:c.update},{default:e(()=>[s(Q,{class:"zone-cp-collection","data-testid":"zone-cp-collection",headers:[{...z.get("headers.type"),label:" ",key:"type"},{...z.get("headers.name"),label:"Name",key:"name"},{...z.get("headers.zoneCpVersion"),label:"Zone Leader CP Version",key:"zoneCpVersion"},{...z.get("headers.ingress"),label:"Ingresses (online / total)",key:"ingress"},{...z.get("headers.egress"),label:"Egresses (online / total)",key:"egress"},{...z.get("headers.state"),label:"Status",key:"state"},{...z.get("headers.warnings"),label:"Warnings",key:"warnings",hideLabel:!0},{...z.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],items:l==null?void 0:l.items,"is-selected-row":o=>o.name===c.params.zone,onResize:z.set},{type:e(({row:o})=>[(r(!0),y(w,null,v([["kubernetes","universal"].find(t=>t===o.zoneInsight.environment)??"kubernetes"],t=>(r(),_(S,{key:t,name:t},{default:e(()=>[a(p(g(`common.product.environment.${t}`)),1)]),_:2},1032,["name"]))),128))]),name:e(({row:o})=>[s(f,{"data-action":"",to:{name:"zone-cp-detail-view",params:{zone:o.name},query:{page:c.params.page,size:c.params.size}}},{default:e(()=>[a(p(o.name),1)]),_:2},1032,["to"])]),zoneCpVersion:e(({row:o})=>[a(p(R(U)(o.zoneInsight,"version.kumaCp.version",g("common.collection.none"))),1)]),ingress:e(({row:o})=>[(r(!0),y(w,null,v([V.value[o.name]||{online:[],offline:[]}],t=>(r(),y(w,null,[a(p(t.online.length)+" / "+p(t.online.length+t.offline.length),1)],64))),256))]),egress:e(({row:o})=>[(r(!0),y(w,null,v([X.value[o.name]||{online:[],offline:[]}],t=>(r(),y(w,null,[a(p(t.online.length)+" / "+p(t.online.length+t.offline.length),1)],64))),256))]),state:e(({row:o})=>[s(Y,{status:o.state},null,8,["status"])]),warnings:e(({row:o})=>[o.warnings.length>0?(r(),_(S,{key:0,name:"warning","data-testid":"warning"},{default:e(()=>[L("ul",null,[(r(!0),y(w,null,v(o.warnings,t=>(r(),y("li",{key:t.kind,"data-testid":`warning-${t.kind}`},p(g(`zone-cps.list.${t.kind}`)),9,te))),128))])]),_:2},1024)):(r(),y(w,{key:1},[a(p(g("common.collection.none")),1)],64))]),actions:e(({row:o})=>[s(Z,null,{default:e(()=>[s(B,null,{default:e(({expanded:t,toggle:h})=>[s(f,{to:{name:"zone-cp-detail-view",params:{zone:o.name}}},{default:e(()=>[a(p(g("common.collection.actions.view")),1)]),_:2},1032,["to"]),n[2]||(n[2]=a()),D("create zones")?(r(),_(f,{key:0,appearance:"danger",onClick:h},{default:e(()=>[a(p(g("common.collection.actions.delete")),1)]),_:2},1032,["onClick"])):b("",!0),n[3]||(n[3]=a()),s(k,{to:{name:"modal-layer"}},{default:e(()=>[t?(r(),_(x,{key:0,src:`/zone-cps/${o.name}/delete`,onChange:()=>{h(),j()}},{default:e(({submit:H,error:J})=>[s($,{action:g("common.delete_modal.proceed_button"),expected:o.name,"data-testid":"delete-zone-modal",onCancel:h,onSubmit:()=>H({})},{title:e(()=>[a(p(g("common.delete_modal.title",{type:"Zone"})),1)]),default:e(()=>[n[0]||(n[0]=a()),s(m,{path:"common.delete_modal.text",params:{type:"Zone",name:o.name}},null,8,["params"]),n[1]||(n[1]=a()),s(I,{class:"mt-4",errors:[J],loader:!1},null,8,["errors"])]),_:2},1032,["action","expected","onCancel","onSubmit"])]),_:2},1032,["src","onChange"])):b("",!0)]),_:2},1024)]),_:2},1024)]),_:2},1024)]),_:2},1032,["headers","items","is-selected-row","onResize"])]),_:2},1032,["items","page","page-size","total","onChange"])]),_:2},1032,["data","errors"])]),_:2},1024),n[15]||(n[15]=a()),c.params.zone?(r(),_(P,{key:0},{default:e(o=>[s(oe,{onClose:t=>c.replace({name:"zone-cp-list-view",query:{page:c.params.page,size:c.params.size}})},{default:e(()=>[(r(),_(ee(o.Component),{name:c.params.zone,"zone-overview":l==null?void 0:l.items.find(t=>t.name===c.params.zone)},null,8,["name","zone-overview"]))]),_:2},1032,["onClose"])]),_:2},1024)):b("",!0)]),_:2},1032,["src"])]),_:2},1032,["docs"])]),_:1})}}}),re=ne(se,[["__scopeId","data-v-dddb6f43"]]);export{re as default};

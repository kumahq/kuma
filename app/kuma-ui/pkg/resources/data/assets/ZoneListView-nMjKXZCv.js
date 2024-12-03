import{d as U,v as T,e as r,o as i,p as _,w as e,a,l as v,b as s,m as I,C as W,t as d,q as b,A as j,c as y,J as w,K as h,U as O,S as Q,F as Y,_ as ee}from"./index-CFsM3b-2.js";import{S as ne}from"./SummaryView-B9clmhmE.js";const oe=["innerHTML"],te=["data-testid"],se=["innerHTML"],ae=U({__name:"ZoneListView",setup(le){const S=T({}),D=T({}),A=C=>{const n="zoneIngress";S.value=C.items.reduce((u,g)=>{var z;const c=(z=g[n])==null?void 0:z.zone;if(typeof c<"u"){typeof u[c]>"u"&&(u[c]={online:[],offline:[]});const k=typeof g[`${n}Insight`].connectedSubscription<"u"?"online":"offline";u[c][k].push(g)}return u},{})},R=C=>{const n="zoneEgress";D.value=C.items.reduce((u,g)=>{var z;const c=(z=g[n])==null?void 0:z.zone;if(typeof c<"u"){typeof u[c]>"u"&&(u[c]={online:[],offline:[]});const k=typeof g[`${n}Insight`].connectedSubscription<"u"?"online":"offline";u[c][k].push(g)}return u},{})};return(C,n)=>{const u=r("RouteTitle"),g=r("DataSource"),c=r("XAction"),z=r("XTeleportTemplate"),k=r("XIcon"),L=r("DataLoader"),$=r("XPrompt"),x=r("DataSink"),B=r("XDisclosure"),M=r("XActionGroup"),N=r("DataCollection"),Z=r("KCard"),E=r("RouterView"),H=r("AppView"),q=r("RouteView");return i(),_(q,{name:"zone-cp-list-view",params:{page:1,size:50,zone:""}},{default:e(({route:p,t:m,can:X,uri:K,me:f})=>[a(H,{docs:m("zones.href.docs.cta")},{title:e(()=>[v("h1",null,[a(u,{title:m("zone-cps.routes.items.title")},null,8,["title"])])]),default:e(()=>[n[16]||(n[16]=s()),a(g,{src:K(I(W),"/zone-cps",{},{page:p.params.page,size:p.params.size})},{default:e(({data:l,error:P,refresh:F})=>[a(g,{src:"/zone-ingress-overviews?page=1&size=100",onChange:A}),n[12]||(n[12]=s()),a(g,{src:"/zone-egress-overviews?page=1&size=100",onChange:R}),n[13]||(n[13]=s()),v("div",{innerHTML:m("zone-cps.routes.items.intro",{},{defaultMessage:""})},null,8,oe),n[14]||(n[14]=s()),a(Z,null,{default:e(()=>[X("create zones")&&((l==null?void 0:l.items)??[]).length>0?(i(),_(z,{key:0,to:{name:"zone-cp-list-view-actions"}},{default:e(()=>[a(c,{action:"create",appearance:"primary",to:{name:"zone-create-view"},"data-testid":"create-zone-link"},{default:e(()=>[s(d(m("zones.index.create")),1)]),_:2},1024)]),_:2},1024)):b("",!0),n[11]||(n[11]=s()),a(L,{data:[l],errors:[P]},{loadable:e(()=>[a(N,{type:"zone-cps",items:(l==null?void 0:l.items)??[void 0],page:p.params.page,"page-size":p.params.size,total:l==null?void 0:l.total,onChange:p.update},{default:e(()=>[a(j,{class:"zone-cp-collection","data-testid":"zone-cp-collection",headers:[{...f.get("headers.type"),label:" ",key:"type"},{...f.get("headers.name"),label:"Name",key:"name"},{...f.get("headers.zoneCpVersion"),label:"Zone Leader CP Version",key:"zoneCpVersion"},{...f.get("headers.ingress"),label:"Ingresses (online / total)",key:"ingress"},{...f.get("headers.egress"),label:"Egresses (online / total)",key:"egress"},{...f.get("headers.state"),label:"Status",key:"state"},{...f.get("headers.warnings"),label:"Warnings",key:"warnings",hideLabel:!0},{...f.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],items:l==null?void 0:l.items,"is-selected-row":o=>o.name===p.params.zone,onResize:f.set},{type:e(({row:o})=>[(i(!0),y(w,null,h([["kubernetes","universal"].find(t=>t===o.zoneInsight.environment)??"kubernetes"],t=>(i(),_(k,{key:t,name:t},{default:e(()=>[s(d(m(`common.product.environment.${t}`)),1)]),_:2},1032,["name"]))),128))]),name:e(({row:o})=>[a(c,{"data-action":"",to:{name:"zone-cp-detail-view",params:{zone:o.name},query:{page:p.params.page,size:p.params.size}}},{default:e(()=>[s(d(o.name),1)]),_:2},1032,["to"])]),zoneCpVersion:e(({row:o})=>[s(d(I(O)(o.zoneInsight,"version.kumaCp.version",m("common.collection.none"))),1)]),ingress:e(({row:o})=>[(i(!0),y(w,null,h([S.value[o.name]||{online:[],offline:[]}],t=>(i(),y(w,null,[s(d(t.online.length)+" / "+d(t.online.length+t.offline.length),1)],64))),256))]),egress:e(({row:o})=>[(i(!0),y(w,null,h([D.value[o.name]||{online:[],offline:[]}],t=>(i(),y(w,null,[s(d(t.online.length)+" / "+d(t.online.length+t.offline.length),1)],64))),256))]),state:e(({row:o})=>[a(Q,{status:o.state},null,8,["status"])]),warnings:e(({row:o})=>[o.warnings.length>0?(i(),_(k,{key:0,name:"warning","data-testid":"warning"},{default:e(()=>[v("ul",null,[(i(!0),y(w,null,h(o.warnings,t=>(i(),y("li",{key:t.kind,"data-testid":`warning-${t.kind}`},d(m(`zone-cps.list.${t.kind}`)),9,te))),128))])]),_:2},1024)):(i(),y(w,{key:1},[s(d(m("common.collection.none")),1)],64))]),actions:e(({row:o})=>[a(M,null,{default:e(()=>[a(B,null,{default:e(({expanded:t,toggle:V})=>[a(c,{to:{name:"zone-cp-detail-view",params:{zone:o.name}}},{default:e(()=>[s(d(m("common.collection.actions.view")),1)]),_:2},1032,["to"]),n[2]||(n[2]=s()),X("create zones")?(i(),_(c,{key:0,appearance:"danger",onClick:V},{default:e(()=>[s(d(m("common.collection.actions.delete")),1)]),_:2},1032,["onClick"])):b("",!0),n[3]||(n[3]=s()),a(z,{to:{name:"modal-layer"}},{default:e(()=>[t?(i(),_(x,{key:0,src:`/zone-cps/${o.name}/delete`,onChange:()=>{V(),F()}},{default:e(({submit:G,error:J})=>[a($,{action:m("common.delete_modal.proceed_button"),expected:o.name,"data-testid":"delete-zone-modal",onCancel:V,onSubmit:()=>G({})},{title:e(()=>[s(d(m("common.delete_modal.title",{type:"Zone"})),1)]),default:e(()=>[n[0]||(n[0]=s()),v("div",{innerHTML:m("common.delete_modal.text",{type:"Zone",name:o.name})},null,8,se),n[1]||(n[1]=s()),a(L,{class:"mt-4",errors:[J],loader:!1},null,8,["errors"])]),_:2},1032,["action","expected","onCancel","onSubmit"])]),_:2},1032,["src","onChange"])):b("",!0)]),_:2},1024)]),_:2},1024)]),_:2},1024)]),_:2},1032,["headers","items","is-selected-row","onResize"])]),_:2},1032,["items","page","page-size","total","onChange"])]),_:2},1032,["data","errors"])]),_:2},1024),n[15]||(n[15]=s()),p.params.zone?(i(),_(E,{key:0},{default:e(o=>[a(ne,{onClose:t=>p.replace({name:"zone-cp-list-view",query:{page:p.params.page,size:p.params.size}})},{default:e(()=>[(i(),_(Y(o.Component),{name:p.params.zone,"zone-overview":l==null?void 0:l.items.find(t=>t.name===p.params.zone)},null,8,["name","zone-overview"]))]),_:2},1032,["onClose"])]),_:2},1024)):b("",!0)]),_:2},1032,["src"])]),_:2},1032,["docs"])]),_:1})}}}),ce=ee(ae,[["__scopeId","data-v-b2ef2ed0"]]);export{ce as default};

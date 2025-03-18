import{d as O,y as L,m as _,w as e,b as a,r,e as l,p as C,t as p,C as Q,c as y,q as R,v,F as w,S as U,s as N,K as Y,o as i,H as ee,N as ne,_ as oe}from"./index-D_WxlpfD.js";import{S as te}from"./SummaryView-AYz_OLVZ.js";const se=["data-testid"],ae=O({__name:"ZoneListView",setup(le){const X=L({}),S=L({}),T=h=>{const n="zoneIngress";X.value=h.items.reduce((u,d)=>{var f;const m=(f=d[n])==null?void 0:f.zone;if(typeof m<"u"){typeof u[m]>"u"&&(u[m]={online:[],offline:[]});const k=typeof d[`${n}Insight`].connectedSubscription<"u"?"online":"offline";u[m][k].push(d)}return u},{})},$=h=>{const n="zoneEgress";S.value=h.items.reduce((u,d)=>{var f;const m=(f=d[n])==null?void 0:f.zone;if(typeof m<"u"){typeof u[m]>"u"&&(u[m]={online:[],offline:[]});const k=typeof d[`${n}Insight`].connectedSubscription<"u"?"online":"offline";u[m][k].push(d)}return u},{})};return(h,n)=>{const u=r("RouteTitle"),d=r("DataSource"),m=r("XI18n"),f=r("XAction"),k=r("XTeleportTemplate"),D=r("XIcon"),I=r("DataLoader"),x=r("XPrompt"),B=r("DataSink"),Z=r("XDisclosure"),E=r("XActionGroup"),q=r("DataCollection"),P=r("XCard"),F=r("RouterView"),G=r("AppView"),H=r("RouteView");return i(),_(H,{name:"zone-cp-list-view",params:{page:1,size:Number,zone:""}},{default:e(({route:c,t:g,can:b,uri:K,me:z})=>[a(d,{src:K(N(ee),"/zone-cps",{},{page:c.params.page,size:c.params.size})},{default:e(({data:t,error:W,refresh:j})=>{var A;return[a(G,{docs:t&&((A=t==null?void 0:t.items)!=null&&A.length)?g("zones.href.docs.cta"):""},{title:e(()=>[R("h1",null,[a(u,{title:g("zone-cps.routes.items.title")},null,8,["title"])])]),default:e(()=>[n[12]||(n[12]=l()),a(d,{src:"/zone-ingress-overviews?page=1&size=100",onChange:T}),n[13]||(n[13]=l()),a(d,{src:"/zone-egress-overviews?page=1&size=100",onChange:$}),n[14]||(n[14]=l()),!b("view growth-new-empty-states")||t!=null&&t.items.length?(i(),_(m,{key:0,path:"zone-cps.routes.items.intro","default-path":"common.i18n.ignore-error"})):C("",!0),n[15]||(n[15]=l()),a(P,null,{default:e(()=>[b("create zones")&&((t==null?void 0:t.items)??[]).length>0?(i(),_(k,{key:0,to:{name:"zone-cp-list-view-actions"}},{default:e(()=>[a(f,{action:"create",appearance:"primary",to:{name:"zone-create-view"},"data-testid":"create-zone-link"},{default:e(()=>[l(p(g("zones.index.create")),1)]),_:2},1024)]),_:2},1024)):C("",!0),n[11]||(n[11]=l()),a(I,{data:[t],errors:[W]},{loadable:e(()=>[a(q,{type:"zone-cps",items:(t==null?void 0:t.items)??[void 0],page:c.params.page,"page-size":c.params.size,total:t==null?void 0:t.total,onChange:c.update},{default:e(()=>[a(Q,{class:"zone-cp-collection","data-testid":"zone-cp-collection",headers:[{...z.get("headers.type"),label:" ",key:"type"},{...z.get("headers.name"),label:"Name",key:"name"},{...z.get("headers.zoneCpVersion"),label:"Zone Leader CP Version",key:"zoneCpVersion"},{...z.get("headers.ingress"),label:"Ingresses (online / total)",key:"ingress"},{...z.get("headers.egress"),label:"Egresses (online / total)",key:"egress"},{...z.get("headers.state"),label:"Status",key:"state"},{...z.get("headers.warnings"),label:"Warnings",key:"warnings",hideLabel:!0},{...z.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],items:t==null?void 0:t.items,"is-selected-row":o=>o.name===c.params.zone,onResize:z.set},{type:e(({row:o})=>[(i(!0),y(w,null,v([["kubernetes","universal"].find(s=>s===o.zoneInsight.environment)??"kubernetes"],s=>(i(),_(D,{key:s,name:s},{default:e(()=>[l(p(g(`common.product.environment.${s}`)),1)]),_:2},1032,["name"]))),128))]),name:e(({row:o})=>[a(f,{"data-action":"",to:{name:"zone-cp-detail-view",params:{zone:o.name},query:{page:c.params.page,size:c.params.size}}},{default:e(()=>[l(p(o.name),1)]),_:2},1032,["to"])]),zoneCpVersion:e(({row:o})=>[l(p(N(ne)(o.zoneInsight,"version.kumaCp.version",g("common.collection.none"))),1)]),ingress:e(({row:o})=>[(i(!0),y(w,null,v([X.value[o.name]||{online:[],offline:[]}],s=>(i(),y(w,null,[l(p(s.online.length)+" / "+p(s.online.length+s.offline.length),1)],64))),256))]),egress:e(({row:o})=>[(i(!0),y(w,null,v([S.value[o.name]||{online:[],offline:[]}],s=>(i(),y(w,null,[l(p(s.online.length)+" / "+p(s.online.length+s.offline.length),1)],64))),256))]),state:e(({row:o})=>[a(U,{status:o.state},null,8,["status"])]),warnings:e(({row:o})=>[o.warnings.length>0?(i(),_(D,{key:0,name:"warning","data-testid":"warning"},{default:e(()=>[R("ul",null,[(i(!0),y(w,null,v(o.warnings,s=>(i(),y("li",{key:s.kind,"data-testid":`warning-${s.kind}`},p(g(`zone-cps.list.${s.kind}`)),9,se))),128))])]),_:2},1024)):(i(),y(w,{key:1},[l(p(g("common.collection.none")),1)],64))]),actions:e(({row:o})=>[a(E,null,{default:e(()=>[a(Z,null,{default:e(({expanded:s,toggle:V})=>[a(f,{to:{name:"zone-cp-detail-view",params:{zone:o.name}}},{default:e(()=>[l(p(g("common.collection.actions.view")),1)]),_:2},1032,["to"]),n[2]||(n[2]=l()),b("create zones")?(i(),_(f,{key:0,appearance:"danger",onClick:V},{default:e(()=>[l(p(g("common.collection.actions.delete")),1)]),_:2},1032,["onClick"])):C("",!0),n[3]||(n[3]=l()),a(k,{to:{name:"modal-layer"}},{default:e(()=>[s?(i(),_(B,{key:0,src:`/zone-cps/${o.name}/delete`,onChange:()=>{V(),j()}},{default:e(({submit:J,error:M})=>[a(x,{action:g("common.delete_modal.proceed_button"),expected:o.name,"data-testid":"delete-zone-modal",onCancel:V,onSubmit:()=>J({})},{title:e(()=>[l(p(g("common.delete_modal.title",{type:"Zone"})),1)]),default:e(()=>[n[0]||(n[0]=l()),a(m,{path:"common.delete_modal.text",params:{type:"Zone",name:o.name}},null,8,["params"]),n[1]||(n[1]=l()),a(I,{class:"mt-4",errors:[M],loader:!1},null,8,["errors"])]),_:2},1032,["action","expected","onCancel","onSubmit"])]),_:2},1032,["src","onChange"])):C("",!0)]),_:2},1024)]),_:2},1024)]),_:2},1024)]),_:2},1032,["headers","items","is-selected-row","onResize"])]),_:2},1032,["items","page","page-size","total","onChange"])]),_:2},1032,["data","errors"])]),_:2},1024),n[16]||(n[16]=l()),c.params.zone?(i(),_(F,{key:1},{default:e(o=>[a(te,{onClose:s=>c.replace({name:"zone-cp-list-view",query:{page:c.params.page,size:c.params.size}})},{default:e(()=>[(i(),_(Y(o.Component),{name:c.params.zone,"zone-overview":t==null?void 0:t.items.find(s=>s.name===c.params.zone)},null,8,["name","zone-overview"]))]),_:2},1032,["onClose"])]),_:2},1024)):C("",!0)]),_:2},1032,["docs"])]}),_:2},1032,["src"])]),_:1})}}}),ce=oe(ae,[["__scopeId","data-v-1ed7965f"]]);export{ce as default};

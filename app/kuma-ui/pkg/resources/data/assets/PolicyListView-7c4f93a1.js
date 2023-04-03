import{d as H,o as p,k as E,w as o,a as S,u as a,b as Y,e as r,E as K,l as Z,C as ee,m as ae,i as te,r as l,j as M,n as z,P as B,t as se,J as le,c as V,g as n,X as ne,M as oe,x as w,ai as re,a9 as O,y as T,F as ie,z as ue,N as ce,O as pe,aj as me,H as de}from"./index-0be248c4.js";import{_ as ye}from"./PolicyConnections.vue_vue_type_script_setup_true_lang-0af3feeb.js";import{D as ve}from"./DataOverview-f7d22b1c.js";import{_ as fe}from"./LabelList.vue_vue_type_style_index_0_lang-c5bb0283.js";import{T as he}from"./TabsWidget-1980411d.js";import{_ as _e}from"./YamlView.vue_vue_type_script_setup_true_lang-9e6ca5fd.js";import{Q as F}from"./QueryParameter-70743f73.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-12f1a66a.js";import"./ErrorBlock-f4ceb6b7.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang-e65555f3.js";import"./TagList-777ca8e8.js";import"./StatusBadge-f388e2f2.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-82aa021d.js";import"./toYaml-4e00099e.js";const ge=H({__name:"DocumentationLink",props:{href:{type:String,required:!0}},setup(m){const i=m;return(D,q)=>(p(),E(a(K),{class:"docs-link",appearance:"outline",target:"_blank",to:i.href},{default:o(()=>[S(a(Y),{icon:"externalLink",color:"currentColor",size:"16","hide-title":""}),r(`

    Documentation
  `)]),_:1},8,["to"]))}}),be=m=>(ce("data-v-652a6b0e"),m=m(),pe(),m),ke={class:"kcard-stack"},Pe={class:"kcard-border"},we=be(()=>n("p",null,[n("strong",null,"Warning"),r(` This policy is experimental. If you encountered any problem please open an
                  `),n("a",{href:"https://github.com/kumahq/kuma/issues/new/choose",target:"_blank",rel:"noopener noreferrer"},"issue")],-1)),Ee={class:"kcard-border"},Se={class:"entity-heading","data-testid":"policy-single-entity"},xe={"data-testid":"policy-overview-tab"},Te=H({__name:"PolicyListView",props:{selectedPolicyName:{type:String,required:!1,default:null},policyPath:{type:String,required:!0},offset:{type:Number,required:!1,default:0}},setup(m){const i=m,D=Z(),q=ee(),Q=[{hash:"#overview",title:"Overview"},{hash:"#affected-dpps",title:"Affected DPPs"}],W=me(),f=ae(),N=te(),h=l(!0),_=l(!1),d=l(null),y=l(!0),v=l(!1),g=l(!1),x=l(!1),b=l({}),k=l(null),$=l(null),R=l(i.offset),C=l({headers:[{label:"Name",key:"name"},{label:"Type",key:"type"}],data:[]}),u=M(()=>N.state.policyTypesByPath[i.policyPath]),j=M(()=>N.state.policyTypes.map(e=>({label:e.name,value:e.path,selected:e.path===i.policyPath}))),G=M(()=>N.state.policyTypes.filter(e=>(N.state.sidebar.insights.mesh.policies[e.name]??0)===0).map(e=>e.name));z(()=>f.params.mesh,function(){f.name===i.policyPath&&(h.value=!0,_.value=!1,y.value=!0,v.value=!1,g.value=!1,x.value=!1,d.value=null,L(0))}),z(()=>f.query.ns,function(){h.value=!0,_.value=!1,y.value=!0,v.value=!1,g.value=!1,x.value=!1,d.value=null,L(0)}),L(i.offset);async function L(e){R.value=e,F.set("offset",e>0?e:null),h.value=!0,d.value=null;const s=f.query.ns||null,t=f.params.mesh,P=u.value.path;try{let c;if(t!==null&&s!==null)c=[await D.getSinglePolicyEntity({mesh:t,path:P,name:s})],$.value=null;else{const A={size:B,offset:e},I=await D.getAllPolicyEntitiesFromMesh({mesh:t,path:P},A);c=I.items??[],$.value=I.next}if(c.length>0){C.value.data=c.map(I=>X(I)),x.value=!1,_.value=!1;const A=i.selectedPolicyName??c[0].name;await U({name:A,mesh:t,path:P})}else C.value.data=[],x.value=!0,_.value=!0,v.value=!0}catch(c){c instanceof Error?d.value=c:console.error(c),_.value=!0}finally{h.value=!1,y.value=!1}}function J(e){W.push({name:"policy",params:{...f.params,policyPath:e.value}})}function X(e){if(!e.mesh)return e;const s=e,t={name:"mesh-detail-view",params:{mesh:e.mesh}};return s.meshRoute=t,s}async function U(e){g.value=!1,y.value=!0,v.value=!1;try{const s=await D.getSinglePolicyEntity({mesh:e.mesh,path:u.value.path,name:e.name});if(s){const t=["type","name","mesh"];b.value=se(s,t),F.set("policy",b.value.name),k.value=le(s)}else b.value={},v.value=!0}catch(s){g.value=!0,console.error(s)}finally{y.value=!1}}return(e,s)=>a(u)?(p(),V("div",{key:0,class:O(["relative",a(u).path])},[n("div",ke,[n("div",Pe,[a(u).isExperimental?(p(),E(a(oe),{key:0,"border-variant":"noBorder",class:"mb-4"},{body:o(()=>[S(a(ne),{appearance:"warning"},{alertMessage:o(()=>[we]),_:1})]),_:1})):w("",!0),r(),S(ve,{"selected-entity-name":b.value.name,"page-size":a(B),error:d.value,"is-loading":h.value,"empty-state":{title:"No Data",message:`There are no ${a(u).name} policies present.`},"table-data":C.value,"table-data-is-empty":x.value,next:$.value,"page-offset":R.value,onTableAction:U,onLoadData:L},{additionalControls:o(()=>[S(a(re),{label:"Policies",items:a(j),"label-attributes":{class:"visually-hidden"},appearance:"select","enable-filtering":!0,onSelected:J},{"item-template":o(({item:t})=>[n("span",{class:O({"policy-type-empty":a(G).includes(t.label)})},T(t.label),3)]),_:1},8,["items"]),r(),S(ge,{href:`${a(q)("KUMA_DOCS_URL")}/policies/${a(u).path}/?${a(q)("KUMA_UTM_QUERY_PARAMS")}`,"data-testid":"policy-documentation-link"},null,8,["href"]),r(),e.$route.query.ns?(p(),E(a(K),{key:0,class:"back-button",appearance:"primary",icon:"arrowLeft",to:{name:"policy",params:{policyPath:i.policyPath}}},{default:o(()=>[r(`
              View all
            `)]),_:1},8,["to"])):w("",!0)]),default:o(()=>[r(`
          >
          `)]),_:1},8,["selected-entity-name","page-size","error","is-loading","empty-state","table-data","table-data-is-empty","next","page-offset"])]),r(),n("div",Ee,[_.value===!1?(p(),E(he,{key:0,"has-error":d.value!==null,error:d.value,"is-loading":h.value,tabs:Q},{tabHeader:o(()=>[n("h1",Se,T(a(u).name)+": "+T(b.value.name),1)]),overview:o(()=>[S(fe,{"has-error":g.value,"is-loading":y.value,"is-empty":v.value},{default:o(()=>[n("div",xe,[n("ul",null,[(p(!0),V(ie,null,ue(b.value,(t,P)=>(p(),V("li",{key:P},[n("h4",null,T(P),1),r(),n("p",null,T(t),1)]))),128))])])]),_:1},8,["has-error","is-loading","is-empty"]),r(),k.value!==null?(p(),E(_e,{key:0,id:"code-block-policy",class:"mt-4","has-error":g.value,"is-loading":y.value,"is-empty":v.value,content:k.value,"is-searchable":""},null,8,["has-error","is-loading","is-empty","content"])):w("",!0)]),"affected-dpps":o(()=>[k.value!==null?(p(),E(ye,{key:0,mesh:k.value.mesh,"policy-name":k.value.name,"policy-type":a(u).path},null,8,["mesh","policy-name","policy-type"])):w("",!0)]),_:1},8,["has-error","error","is-loading"])):w("",!0)])])],2)):w("",!0)}});const Oe=de(Te,[["__scopeId","data-v-652a6b0e"]]);export{Oe as default};

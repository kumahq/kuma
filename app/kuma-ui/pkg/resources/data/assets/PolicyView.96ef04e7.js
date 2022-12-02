import{d as F,o as r,c as C,w as l,a as w,u as s,v as Q,b as T,M as U,r as e,e as $,f as O,g as Z,k as L,h as J,i as b,j as t,l as X,m as ee,F as H,n as W,t as N,p as ae,q as te,P as R,s as se,x as le,y as ne,z as E,A as oe,B as re,C as ie,D as ue}from"./index.3bc39668.js";import{p as ce,D as pe}from"./patchQueryParam.65a1b943.js";import{F as me}from"./FrameSkeleton.e1893be2.js";import{_ as Y}from"./LabelList.vue_vue_type_style_index_0_lang.0e14ac31.js";import{T as de}from"./TabsWidget.1751eed8.js";import{Y as ve}from"./YamlView.24c9d3cb.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang.74b6b406.js";import"./ErrorBlock.f4ac98cc.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang.13b03cfc.js";import"./TagList.3d4ee64d.js";import"./index.58caa11d.js";import"./CodeBlock.vue_vue_type_style_index_0_lang.b3d38a49.js";import"./_commonjsHelpers.f037b798.js";const ye=F({__name:"DocumentationLink",props:{href:{type:String,required:!0}},setup(d){const i=d;return(S,v)=>(r(),C(s(U),{class:"docs-link",appearance:"outline",target:"_blank",to:i.href},{icon:l(()=>[w(s(Q),{icon:"externalLink"})]),default:l(()=>[T(" Documentation ")]),_:1},8,["to"]))}}),fe=t("h4",null,"Dataplanes",-1),he=F({__name:"PolicyConnections",props:{mesh:{type:String,required:!0},policyType:{type:String,required:!0},policyName:{type:String,required:!0}},setup(d){const i=d,S=e(!1),v=e(!0),D=e(!1),_=e([]),y=e(""),g=$(()=>{const u=y.value.toLowerCase();return _.value.filter(({dataplane:n})=>n.name.toLowerCase().includes(u))});O(()=>i.policyName,function(){f()}),Z(function(){f()});async function f(){D.value=!1,v.value=!0;try{const{items:u,total:n}=await L.getPolicyConnections({mesh:i.mesh,policyType:i.policyType,policyName:i.policyName});S.value=n>0,_.value=u}catch{D.value=!0}finally{v.value=!1}}return(u,n)=>{const P=J("router-link");return r(),b("div",null,[w(Y,{"has-error":D.value,"is-loading":v.value,"is-empty":!S.value},{default:l(()=>[t("ul",null,[t("li",null,[fe,X(t("input",{id:"dataplane-search","onUpdate:modelValue":n[0]||(n[0]=c=>y.value=c),type:"text",class:"k-input mb-4",placeholder:"Filter by name",required:"","data-testid":"dataplane-search-input"},null,512),[[ee,y.value]]),(r(!0),b(H,null,W(s(g),(c,k)=>(r(),b("p",{key:k,class:"my-1","data-testid":"dataplane-name"},[w(P,{to:{name:"data-plane-detail-view",params:{mesh:c.dataplane.mesh,dataPlane:c.dataplane.name}}},{default:l(()=>[T(N(c.dataplane.name),1)]),_:2},1032,["to"])]))),128))])])]),_:1},8,["has-error","is-loading","is-empty"])])}}}),j=d=>(re("data-v-1087084f"),d=d(),ie(),d),_e={key:0,class:"mb-4"},ge=j(()=>t("p",null,[t("strong",null,"Warning"),T(" This policy is experimental. If you encountered any problem please open an "),t("a",{href:"https://github.com/kumahq/kuma/issues/new/choose",target:"_blank",rel:"noopener noreferrer"},"issue")],-1)),ke=j(()=>t("span",{class:"custom-control-icon"}," \u2190 ",-1)),we={"data-testid":"policy-single-entity"},be={"data-testid":"policy-overview-tab"},De={class:"config-wrapper"},Pe=F({__name:"PolicyView",props:{policyPath:{type:String,required:!0},offset:{type:Number,required:!1,default:0}},setup(d){const i=d,S=[{hash:"#overview",title:"Overview"},{hash:"#affected-dpps",title:"Affected DPPs"}],v=ae(),D=te(),_=e(!0),y=e(!1),g=e(null),f=e(!0),u=e(!1),n=e(!1),P=e(!1),c=e({}),k=e(null),q=e(null),B=e(i.offset),I=e({headers:[{label:"Actions",key:"actions",hideLabel:!0},{label:"Name",key:"name"},{label:"Type",key:"type"}],data:[]}),p=$(()=>D.state.policiesByPath[i.policyPath]),G=$(()=>`https://kuma.io/docs/${D.getters["config/getKumaDocsVersion"]}/policies/${p.value.path}/`);O(()=>v.params.mesh,function(){v.name===i.policyPath&&(_.value=!0,y.value=!1,f.value=!0,u.value=!1,n.value=!1,P.value=!1,g.value=null,A(0))}),A(i.offset);async function A(a){B.value=a,ce("offset",a>0?a:null),_.value=!0,g.value=null;const o=v.query.ns||null,h=v.params.mesh,x=p.value.path;try{let m;if(h!==null&&o!==null)m=[await L.getSinglePolicyEntity({mesh:h,path:x,name:o})],q.value=null;else{const V={size:R,offset:a},z=await L.getAllPolicyEntitiesFromMesh({mesh:h,path:x},V);m=z.items,q.value=z.next}m.length>0?(I.value.data=m.map(V=>K(V)),P.value=!1,y.value=!1,await M({mesh:m[0].mesh,name:m[0].name,path:x})):(I.value.data=[],P.value=!0,y.value=!0,u.value=!0)}catch(m){m instanceof Error?g.value=m:console.error(m),y.value=!0}finally{_.value=!1,f.value=!1}}function K(a){if(!a.mesh)return a;const o=a,h={name:"mesh-detail-view",params:{mesh:a.mesh}};return o.meshRoute=h,o}async function M(a){n.value=!1,f.value=!0,u.value=!1;try{const o=await L.getSinglePolicyEntity({mesh:a.mesh,path:p.value.path,name:a.name});if(o){const h=["type","name","mesh"];c.value=se(o,h),k.value=le(o)}else c.value={},u.value=!0}catch(o){n.value=!0,console.error(o)}finally{f.value=!1}}return(a,o)=>s(p)?(r(),b("div",{key:0,class:oe(["relative",s(p).path])},[s(p).isExperimental?(r(),b("div",_e,[w(s(ne),{appearance:"warning"},{alertMessage:l(()=>[ge]),_:1})])):E("",!0),w(me,null,{default:l(()=>[w(pe,{"selected-entity-name":c.value.name,"page-size":s(R),error:g.value,"is-loading":_.value,"empty-state":{title:"No Data",message:`There are no ${s(p).pluralDisplayName} present.`},"table-data":I.value,"table-data-is-empty":P.value,next:q.value,"page-offset":B.value,onTableAction:M,onLoadData:A},{additionalControls:l(()=>[w(ye,{href:s(G),"data-testid":"policy-documentation-link"},null,8,["href"]),a.$route.query.ns?(r(),C(s(U),{key:0,class:"back-button",appearance:"primary",to:{name:s(p).path}},{default:l(()=>[ke,T(" View All ")]),_:1},8,["to"])):E("",!0)]),default:l(()=>[T(" > ")]),_:1},8,["selected-entity-name","page-size","error","is-loading","empty-state","table-data","table-data-is-empty","next","page-offset"]),y.value===!1?(r(),C(de,{key:0,"has-error":g.value!==null,error:g.value,"is-loading":_.value,tabs:S,"initial-tab-override":"overview"},{tabHeader:l(()=>[t("div",null,[t("h3",we,N(s(p).singularDisplayName)+": "+N(c.value.name),1)])]),overview:l(()=>[w(Y,{"has-error":n.value,"is-loading":f.value,"is-empty":u.value},{default:l(()=>[t("div",be,[t("ul",null,[(r(!0),b(H,null,W(c.value,(h,x)=>(r(),b("li",{key:x},[t("h4",null,N(x),1),t("p",null,N(h),1)]))),128))])])]),_:1},8,["has-error","is-loading","is-empty"]),t("div",De,[k.value!==null?(r(),C(ve,{key:0,id:"code-block-policy","has-error":n.value,"is-loading":f.value,"is-empty":u.value,content:k.value,"is-searchable":""},null,8,["has-error","is-loading","is-empty","content"])):E("",!0)])]),"affected-dpps":l(()=>[k.value!==null?(r(),C(he,{key:0,mesh:k.value.mesh,"policy-name":k.value.name,"policy-type":s(p).path},null,8,["mesh","policy-name","policy-type"])):E("",!0)]),_:1},8,["has-error","error","is-loading"])):E("",!0)]),_:1})],2)):E("",!0)}});const Be=ue(Pe,[["__scopeId","data-v-1087084f"]]);export{Be as default};

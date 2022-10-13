import{d as $,o,c as I,w as s,a as v,b as V,r as L,_ as M,K as S,e as D,f as t,g as Q,v as X,F as H,h as O,t as T,u as ee,i,j as z,k as te,P as F,l as R,s as U,m as y,n as E,p as ae,q as se,x as le,y as ne}from"./index.04875eef.js";import{D as oe}from"./DataOverview.cb1ff33f.js";import{E as re}from"./EntityURLControl.4f8a12bf.js";import{F as ie}from"./FrameSkeleton.49389b31.js";import{L as W}from"./LabelList.ba2075c7.js";import{T as ue}from"./TabsWidget.0ad3f108.js";import{Y as ce}from"./YamlView.eaa3134f.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang.0fcd7b43.js";import"./ErrorBlock.08062bd2.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang.4361c91d.js";import"./index.58caa11d.js";import"./CodeBlock.58f32651.js";import"./_commonjsHelpers.712cc82f.js";const pe=$({__name:"DocumentationLink",props:{href:{type:String,required:!0}},setup(a){return(u,K)=>{const g=L("KIcon"),c=L("KButton");return o(),I(c,{class:"docs-link",appearance:"outline",target:"_blank",to:a.href},{icon:s(()=>[v(g,{icon:"externalLink"})]),default:s(()=>[V(" Documentation ")]),_:1},8,["to"])}}}),me={name:"PolicyConnections",components:{LabelList:W},props:{mesh:{type:String,required:!0},policyType:{type:String,required:!0},policyName:{type:String,required:!0}},data(){return{hasDataplanes:!1,isLoading:!0,hasError:!1,dataplanes:[],searchInput:""}},computed:{filteredDataplanes(){const a=this.searchInput.toLowerCase();return this.dataplanes.filter(({dataplane:{name:u}})=>u.toLowerCase().includes(a))}},watch:{policyName(){this.fetchPolicyConntections()}},mounted(){this.fetchPolicyConntections()},methods:{async fetchPolicyConntections(){this.hasError=!1,this.isLoading=!0;try{const{items:a,total:u}=await S.getPolicyConnections({mesh:this.mesh,policyType:this.policyType,policyName:this.policyName});this.hasDataplanes=u>0,this.dataplanes=a}catch{this.hasError=!0}finally{this.isLoading=!1}}}},de=t("h4",null,"Dataplanes",-1);function he(a,u,K,g,c,b){const k=L("router-link"),d=L("LabelList");return o(),D("div",null,[v(d,{"has-error":c.hasError,"is-loading":c.isLoading,"is-empty":!c.hasDataplanes},{default:s(()=>[t("ul",null,[t("li",null,[de,Q(t("input",{id:"dataplane-search","onUpdate:modelValue":u[0]||(u[0]=l=>c.searchInput=l),type:"text",class:"k-input mb-4",placeholder:"Filter by name",required:""},null,512),[[X,c.searchInput]]),(o(!0),D(H,null,O(b.filteredDataplanes,(l,f)=>(o(),D("p",{key:f,class:"my-1","data-testid":"dataplane-name"},[v(k,{to:{name:"data-plane-list-view",query:{ns:l.dataplane.name},params:{mesh:l.dataplane.mesh}}},{default:s(()=>[V(T(l.dataplane.name),1)]),_:2},1032,["to"])]))),128))])])]),_:1},8,["has-error","is-loading","is-empty"])])}const ye=M(me,[["render",he]]),Y=a=>(le("data-v-126adf33"),a=a(),ne(),a),ve={key:0,class:"mb-4"},fe=Y(()=>t("p",null,[t("strong",null,"Warning"),V(" This policy is experimental. If you encountered any problem please open an "),t("a",{href:"https://github.com/kumahq/kuma/issues/new/choose",target:"_blank",rel:"noopener noreferrer"},"issue")],-1)),_e=Y(()=>t("span",{class:"custom-control-icon"}," \u2190 ",-1)),ge={"data-testid":"policy-single-entity"},be={"data-testid":"policy-overview-tab"},ke={class:"config-wrapper"},we=$({__name:"PolicyView",props:{policyPath:{type:String,required:!0}},setup(a){const u=a,K=[{hash:"#overview",title:"Overview"},{hash:"#affected-dpps",title:"Affected DPPs"}],g=ee(),c=se(),b=i(!0),k=i(!1),d=i(null),l=i(!0),f=i(!1),P=i(!1),N=i(!1),x=i({}),w=i(null),q=i(null),A=i({headers:[{label:"Actions",key:"actions",hideLabel:!0},{label:"Name",key:"name"},{label:"Mesh",key:"mesh"},{label:"Type",key:"type"}],data:[]}),p=z(()=>c.state.policiesByPath[u.policyPath]),j=z(()=>`https://kuma.io/docs/${c.getters["config/getKumaDocsVersion"]}/policies/${p.value.path}/`);te(()=>g.params.mesh,function(){g.name===u.policyPath&&(b.value=!0,k.value=!1,l.value=!0,f.value=!1,P.value=!1,N.value=!1,d.value=null,B())}),B();async function B(e=0){b.value=!0,d.value=null;const n=g.query.ns||null,r=g.params.mesh||null,C=p.value.path;try{let m;if(r!==null&&n!==null)m=[await S.getSinglePolicyEntity({mesh:r,path:C,name:n})],q.value=null;else if(r===null||r==="all"){const h={size:F,offset:e},_=await S.getAllPolicyEntities({path:C},h);m=_.items,q.value=_.next}else{const h={size:F,offset:e},_=await S.getAllPolicyEntitiesFromMesh({mesh:r,path:C},h);m=_.items,q.value=_.next}if(m.length>0){A.value.data=m.map(J=>G(J)),N.value=!1,k.value=!1;const h=["type","name","mesh"],_=m[0];x.value=R(_,h),w.value=U(_)}else A.value.data=[],N.value=!0,k.value=!0,f.value=!0}catch(m){d.value=m,k.value=!0}finally{b.value=!1,l.value=!1}}function G(e){if(!e.mesh)return e;const n=e,r={name:"mesh-child",params:{mesh:e.mesh}};return n.meshRoute=r,n}async function Z(e){if(P.value=!1,l.value=!0,f.value=!1,e)try{const n=await S.getSinglePolicyEntity({mesh:e.mesh,path:p.value.path,name:e.name});if(n){const r=["type","name","mesh"];e.value=R(n,r),w.value=U(n)}else e.value={},f.value=!0}catch(n){P.value=!0,console.error(n)}finally{l.value=!1}}return(e,n)=>{const r=L("KAlert"),C=L("KButton");return y(p)?(o(),D("div",{key:0,class:ae(["relative",y(p).path])},[y(p).isExperimental?(o(),D("div",ve,[v(r,{appearance:"warning"},{alertMessage:s(()=>[fe]),_:1})])):E("",!0),v(ie,null,{default:s(()=>[v(oe,{"page-size":y(F),"has-error":d.value!==null,error:d.value,"is-loading":b.value,"empty-state":{title:"No Data",message:`There are no ${y(p).pluralDisplayName} present.`},"table-data":A.value,"table-data-is-empty":N.value,next:q.value,onTableAction:Z,onLoadData:B},{additionalControls:s(()=>[v(pe,{href:y(j),"data-testid":"policy-documentation-link"},null,8,["href"]),e.$route.query.ns?(o(),I(C,{key:0,class:"back-button",appearance:"primary",to:{name:y(p).path}},{default:s(()=>[_e,V(" View All ")]),_:1},8,["to"])):E("",!0)]),default:s(()=>[V(" > ")]),_:1},8,["page-size","has-error","error","is-loading","empty-state","table-data","table-data-is-empty","next"]),k.value===!1?(o(),I(ue,{key:0,"has-error":d.value!==null,error:d.value,"is-loading":b.value,tabs:K,"initial-tab-override":"overview"},{tabHeader:s(()=>[t("div",null,[t("h3",ge,T(y(p).singularDisplayName)+": "+T(x.value.name),1)]),t("div",null,[v(re,{name:x.value.name,mesh:x.value.mesh},null,8,["name","mesh"])])]),overview:s(()=>[v(W,{"has-error":P.value,"is-loading":l.value,"is-empty":f.value},{default:s(()=>[t("div",be,[t("ul",null,[(o(!0),D(H,null,O(x.value,(m,h)=>(o(),D("li",{key:h},[t("h4",null,T(h),1),t("p",null,T(m),1)]))),128))])])]),_:1},8,["has-error","is-loading","is-empty"]),t("div",ke,[w.value!==null?(o(),I(ce,{key:0,id:"code-block-policy","has-error":P.value,"is-loading":l.value,"is-empty":f.value,content:w.value,"is-searchable":""},null,8,["has-error","is-loading","is-empty","content"])):E("",!0)])]),"affected-dpps":s(()=>[w.value!==null?(o(),I(ye,{key:0,mesh:w.value.mesh,"policy-name":w.value.name,"policy-type":y(p).path},null,8,["mesh","policy-name","policy-type"])):E("",!0)]),_:1},8,["has-error","error","is-loading"])):E("",!0)]),_:1})],2)):E("",!0)}}});const Ae=M(we,[["__scopeId","data-v-126adf33"]]);export{Ae as default};
